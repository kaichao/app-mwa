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
    DIR_RFI="/local${LOCAL_OUTPUT_ROOT}/mwa/rfi"
else
    DIR_RFI=/cluster_data_root/mwa/rfi
fi

source /app/bin/module.env

m=$1
# f_dir=${m}.fits
full_dir="$DIR_FITS/${m}"
echo "file dir:${full_dir}"

# first uncompress the data in-situ if needed.

if [ $COMPRESSED_INPUT = "yes" ]; then

    echo '"before decompress, ls $zst_file*' >> ${WORK_DIR}/custom-out.txt
    ls -l ${full_dir} >> ${WORK_DIR}/custom-out.txt
    cd $full_dir
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
fi

# now all data have been uncompressed.
bname=$m

echo "DIR_RFI:$DIR_RFI/$bname"
mkdir -p $DIR_RFI/$bname
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In mkdir:$bname, ret-code:$code" >&2 && exit 11

cd ${full_dir}
for file in $( ls *.fits )
do
    channel=$(readfile $file | grep "Number of channels" | cut -d "=" -f2 | xargs)
    [[ $channel -ne 3072 ]] && echo "[WARNING] In rfi-find:$full_dir, $file has wrong Num of channels: $channel. Removed $file." >&2 && exit 13
done

cd $DIR_RFI/$bname
echo "RFIARGS: $RFIARGS"
rfifind $RFIARGS -o RFIfile ${full_dir}/*.fits
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In rfi-find:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_RFI/$bname && exit 14

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

if [ "$COMPRESS_OUTFITS" == "yes" ]; then
    cd $full_dir
    for full_name in $( ls *.fits )
    do
        # full_name="$DIR_FITS/${f_dir}"
        zst_file=${full_name}.zst
        echo "full_name:${full_name}" >> ${WORK_DIR}/custom-out.txt
        [ -f "${full_name}" ] && zstd --rm -f ${full_name}

        [[ ! -f $zst_file ]] && echo "[ERROR] In checking file exits:$zst_file, ret-code:$code" >&2 && exit 15
    done
fi

echo $DIR_FITS/${m} >> ${WORK_DIR}/input-files.txt
echo $DIR_RFI/$bname/RFIfile_rfifind.mask >> ${WORK_DIR}/output-files.txt

echo "send message to sink job"
echo ${m} >> ${WORK_DIR}/messages.txt

exit $code