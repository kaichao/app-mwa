#!/bin/bash

# command line args:
# $m: file name to be executed

# environment variables:
# $COMPRESSED_INPUT     the input is compressed fits.zst file
# $NSUB                 nsub for prepsubband_gpu
# $DEDISPARGS           arguments for dedispersion
# $SEARCHARGS           arguments for search
# $LINEMODE             if complete one line of search plan in single execution

# 1. set the input / output / medium file directory

# m="/1257010784/p00017/1"
# source /root/.bashrc
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
if [ $LOCAL_INPUT_ROOT ]; then
    DIR_FITS="/local${LOCAL_INPUT_ROOT}/mwa/24ch"
else
    DIR_FITS=/cluster_data_root/mwa/24ch
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_DEDISP="/local${LOCAL_OUTPUT_ROOT}/mwa/dedisp"
else
    DIR_DEDISP=/cluster_data_root/mwa/dedisp
fi

if [ $PLAN_FILE ]; then
    DDPLAN_FILE=/app/bin/$PLAN_FILE
else
    echo "[ERROR] Search plan not set!" >&2 && exit 10
fi

source /app/bin/module.env

m0=$1
m=${m0%/*}
msgidx=${m0##*/}
dataset=${m%/*}
pointing=${m##*/}
echo "DIR_FITS:$DIR_FITS/$m"
echo "dataset:$dataset"
echo "pointing:$pointing"
# f_dir=${m}.fits
full_dir="$DIR_FITS/${m}"
bname=$m
grpidx=$((10#$msgidx))
# if need to uncompress the file
if [ $COMPRESSED_INPUT = "yes" ]; then

    if [ $grpidx -eq 1 ]; then
        echo '"before decompress, ls $zst_file*' >> ${WORK_DIR}/custom-out.txt
        ls -l ${full_dir} >> ${WORK_DIR}/custom-out.txt
        cd $full_dir
        # this will only execute once.
        for zst_file in $( ls *.zst )
        do
            fits_name=${zst_file%.zst}
            echo "full_name:${fits_name}" >> ${WORK_DIR}/custom-out.txt
            [ -f "${zst_file}" ] && zstd -d --rm -f -o ${fits_name} ${zst_file}
            # 2. check if the file exists
            [[ ! -f $fits_name ]] && echo "[ERROR] In checking file exits:$fits_name, ret-code:$code" >&2 && exit 10
        done
        echo '"after decompress, list all files:' >> ${WORK_DIR}/custom-out.txt
        ls -l ${full_dir} >> ${WORK_DIR}/custom-out.txt
    else
        # check if all files have been decompressed. 
        # if there are still some files not decompressed, wait for the first message to decompress it.
        # get the num of .zst files
        num_zst=$(ls ${full_dir}/*.zst | wc -l)
        if [ $num_zst -gt 0 ]; then
            sleep 120
            # check again.
            num_zst=$(ls ${full_dir}/*.zst | wc -l)
            if [ $num_zst -gt 0 ]; then
                echo "waiting for the first message to decompress the files" >&2
                exit 20
            fi
        fi
    fi
fi

# the file have already been decompressed.
# 3. run the programs to dedisperse and search
echo "DIR_DEDISP:$DIR_DEDISP/$bname"

msgidx=$((10#$msgidx))
echo "message index:${msgidx}"

lines=$(cat $DDPLAN_FILE | wc -l)
echo "lines:$lines"
# if not using LINEMODE
cnt=0
if [ $LINEMODE -eq 0 ]; then
    for ((i=0; i<lines; i++)); do
        # skip the first line
        if [ $i -eq 0 ]; then
            continue
        fi
        line=$(sed -n "$(($i+1))p" $DDPLAN_FILE)
        # get the Ncalls and calls
        NCALLS=$(echo $line | awk '{print $9}')
        calls=$(echo $line | awk '{print $8}')
        # if msgidx < cnt + calls/Ncalls, We find the line number
        if [ $msgidx -le $(($cnt + $calls/$NCALLS)) ]; then
            echo $cnt
            LINENUM=$i
            # set the subgroup number
            SUBGRPNUM=$(( $msgidx - $cnt ))
            break
        fi
        cnt=$(($cnt + $calls/$NCALLS))
    done
# if using LINEMODE, LINENUM is the message index
else
    LINENUM=$msgidx
    SUBGRPNUM=1
    line=$(sed -n "$((LINENUM+1))p" $DDPLAN_FILE)
    # get the Ncalls and calls
    calls=$(echo $line | awk '{print $8}')
    NCALLS=$calls
fi

# get the other parameters from the line
line=$(sed -n "$((LINENUM+1))p" $DDPLAN_FILE)
lodm=$(echo $line | awk '{print $1}')
dmstep=$(echo $line | awk '{print $3}')
downsamp=$(echo $line | awk '{print $4}')
dsubdm=$(echo $line | awk '{print $5}')
dmpercall=$(echo $line | awk '{print $7}')
# update the correct lodm using subgroup number and dsubdm
lodm=$(echo $lodm $dsubdm $SUBGRPNUM $NCALLS | awk '{print $1 + $2 * $4 * ($3 - 1)}')

# get the correct rfi mask path
if [ $DIR_RFI ]; then
    RFI_DIR=/cluster_data_root/mwa/rfi/$DIR_RFI
else
    RFI_DIR=/cluster_data_root/mwa/rfi/$bname
fi
# check if the rfifile exists
if [ ! -f "$RFI_DIR/RFIfile_rfifind.mask" ]; then
    echo "[ERROR] In checking file exits:RFIfile_rfifind.mask, ret-code:$code" >&2
    RFIARGS=""
else
    RFIARGS="-mask $RFI_DIR/RFIfile_rfifind.mask"
fi

LINENUM=$(printf "%02d" "$LINENUM")
mkdir -p ${DIR_DEDISP}/${bname}/dm${LINENUM}/group${msgidx}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In mkdir:dm${LINENUM}/group${msgidx}, ret-code:$code" >&2 && exit 11
cd ${DIR_DEDISP}/${bname}/dm${LINENUM}/group${msgidx}

# run the program for NCALLS times
for ((i=0; i<$NCALLS; i++)); do
    # get the correct DM
    dm=$(echo $lodm $dsubdm $i | awk '{print $1 + $2 * $3}')
    echo "dm:$dm"
    # run the program
    echo "dedisp_search -cuda 0 -ncpus $NCPUS -nsub $NSUB -lodm $dm -dmstep $dmstep -noclip -zerodm \
        -numdms $dmpercall -downsamp $downsamp $RFIARGS -o $pointing $DEDISPARGS $SEARCHARGS $full_dir/*.fits"
    dedisp_search -cuda 0 -ncpus $NCPUS -nsub $NSUB -lodm $dm -dmstep $dmstep -noclip -zerodm \
        -numdms $dmpercall -downsamp $downsamp $RFIARGS -o $pointing $DEDISPARGS $SEARCHARGS $full_dir/*.fits

    code=$?
    echo "dedisp-search ret-code:$code"
    if [ $code -ne 0 ]; then
        echo "[ERROR] In running dedisp-search, ret-code:$code" >&2
        exit 10
    fi
done

echo $DIR_FITS/${m} >> ${WORK_DIR}/input-files.txt
echo $DIR_DEDISP/$bname/dm${LINENUM}/group${msgidx} >> ${WORK_DIR}/output-files.txt

echo "send message to sink job"
echo ${bname}/dm${LINENUM}/group${msgidx} > ${WORK_DIR}/messages.txt
exit $code