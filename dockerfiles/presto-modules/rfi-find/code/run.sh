#!/bin/bash

# 1. set the input / output / medium file directory

# m="1257010784/p00017"
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

if [ $SHARED_ROOT ]; then
    DIR_SHARED="/local${SHARED_ROOT}/mwa/24ch"
fi


m=$1
# f_dir=${m}.fits
full_dir="$DIR_FITS/${m}"
echo "file dir:${full_dir}"

if [ $SHARED_ROOT ]; then
    zst_dir="$DIR_SHARED/${m}"

    echo '"before decompress, ls $zst_file*' >> ${WORK_DIR}/custom-out.txt
    ls -l ${zst_dir} >> ${WORK_DIR}/custom-out.txt
    mkdir -p ${full_dir}
    cd $full_dir
    for zst_file in $( ls ${zst_dir}/*.zst )
    do
        file_name=$(basename $zst_file)
        fits_name=${file_name%.zst}
        echo "full_name:${fits_name}" >> ${WORK_DIR}/custom-out.txt
        [ -f "${zst_file}" ] && zstd -d -f -o ${fits_name} ${zst_file}

    #     # cd $DIR_FITS/$(dirname $1) && [ -f "$(basename $1).fits.zst" ] && zstd -d --rm -f $(basename $1).fits.zst
    #     # 2. check if the file exists

    #     # readfile $DIR_FITS/$f_dir
    #     # code=$?
    #     # [[ $code -ne 0 ]] && echo "[ERROR]Error in checking file exits:$fdir, ret-code:$code" >&2 && exit 10
        [[ ! -f $fits_name ]] && echo "[ERROR] In checking file exits:$fits_name, ret-code:$code" >&2 && exit 10
    done
    echo '"after decompress, list all files:' >> ${WORK_DIR}/custom-out.txt
    ls -l ${full_dir} >> ${WORK_DIR}/custom-out.txt
fi
# get the filename without extension
# arr=($(echo $f_dir | tr "/" "\n"))
# fname=${arr[2]}
# bname=${f_dir%.*}

# all file have already been uncompressed.
bname=$m

# 3. run the programs to dedisperse and search
echo "DIR_DEDISP:$DIR_DEDISP/$bname"
mkdir -p $DIR_DEDISP/$bname
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In mkdir:$bname, ret-code:$code" >&2 && exit 11

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

cd ${full_dir}
for file in $(ls)
do
    channel=$(readfile $file | grep "Number of channels" | cut -d "=" -f2 | xargs)
    [[ $channel -ne 3072 ]] && echo "[WARNING] In rfi-find:$full_dir, $file has wrong Num of channel: $channel. Removed $file." >&2 && rm $file
done

cd $DIR_DEDISP/$bname
rfifind $RFIARGS -o RFIfile ${full_dir}/*.fits
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In rfi-find:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname && rm -rf ${full_dir} && exit 12

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

# for full_name in $( ls ${full_dir}/*.fits )
# do
#     # full_name="$DIR_FITS/${f_dir}"
#     zst_file=${full_name}.zst
#     echo "full_name:${full_name}" >> ${WORK_DIR}/custom-out.txt
#     [ -f "${full_name}" ] && zstd --rm -f ${full_name}

#     [[ ! -f $zst_file ]] && echo "[ERROR] In checking file exits:$zst_file, ret-code:$code" >&2 && exit 13
# done

echo $DIR_FITS/${m} >> ${WORK_DIR}/input-files.txt
echo $DIR_DEDISP/$bname/RFIfile_rfifind.mask >> ${WORK_DIR}/output-files.txt

echo "send message to sink job"
for i in $( seq 1 ${MAX_LINENUM} )
do
    j=$( printf "%02d" "$i" )
    echo ${m}/$j >> ${WORK_DIR}/messages.txt
done

exit $code