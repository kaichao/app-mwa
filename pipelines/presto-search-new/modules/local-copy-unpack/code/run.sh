#!/bin/bash

# similar to pull-unpack
# copy the whole directory to local target directory
# input example: user@url/ssh/dir~1257010784/p00001/t1257010786_1257010936.fits.zst~/target/dir
# input format: dataset/pointing/filename.fits.zst

if [ $LOCAL_INPUT_ROOT ]; then
    DIR_SHARED="/local_data_root${LOCAL_INPUT_ROOT}/mwa/24ch"
else
    DIR_SHARED=/cluster_data_root/mwa/24ch
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_FITS="/local_data_root${LOCAL_OUTPUT_ROOT}/mwa/24ch"
else
    DIR_FITS=/cluster_data_root/mwa/24ch
fi
if [ $BW_LIMIT != "" ]; then
    BW_LIMIT_ARGS="-L $BW_LIMIT"
else
    BW_LIMIT_ARGS=""
fi

m=$1
# f_dir=${m}.fits
full_dir="$DIR_FITS/${m}"
echo "file dir:${full_dir}"

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
    [ -f "${zst_file}" ] && pv $BW_LIMIT_ARGS ${zst_file} | zstd -d -f -o ${fits_name} && touch -a -m ${fits_name}

    [[ ! -f $fits_name ]] && echo "[ERROR] In checking file exits:$fits_name, ret-code:$code" >&2 && exit 10
    [ "$KEEP_SOURCE_FILE" == "no" ] && rm $zst_file
done
echo '"after decompress, list all files:' >> ${WORK_DIR}/custom-out.txt
ls -l ${full_dir} >> ${WORK_DIR}/custom-out.txt

# send message to next module
echo ${m} >> ${WORK_DIR}/messages.txt
