#!/bin/bash

# 1. set the input / output / medium file directory

# m="/1257010784/00017"
# source /root/.bashrc
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
if [ $LOCAL_INPUT_ROOT ]; then
    DIR_FITS="/local${LOCAL_INPUT_ROOT}/mwa/24ch"
else
    DIR_FITS=/data/mwa/24ch
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_DEDISP="/local${LOCAL_OUTPUT_ROOT}/mwa/dedisp"
else
    DIR_DEDISP=/data/mwa/dedisp
fi

# decompress zst file
m=$1
# f_dir=${m}.fits
full_dir="$DIR_FITS/${m}"
echo "file dir:${full_dir}"

echo '"before decompress, ls $zst_file*' >> ${WORK_DIR}/custom-out.txt
ls -l ${full_dir} >> ${WORK_DIR}/custom-out.txt

for zst_file in $( ls ${full_dir}/*.zst )
do
    # full_name="$DIR_FITS/${f_dir}"
    full_name=${zst_file%.zst}
    echo "full_name:${full_name}" >> ${WORK_DIR}/custom-out.txt
    [ -f "${zst_file}" ] && zstd -d --rm -f ${zst_file}

    # cd $DIR_FITS/$(dirname $1) && [ -f "$(basename $1).fits.zst" ] && zstd -d --rm -f $(basename $1).fits.zst
    # 2. check if the file exists

    # readfile $DIR_FITS/$f_dir
    # code=$?
    # [[ $code -ne 0 ]] && echo "[ERROR]Error in checking file exits:$fdir, ret-code:$code" >&2 && exit 10
    [[ ! -f $full_name ]] && echo "[ERROR] In checking file exits:$full_name, ret-code:$code" >&2 && exit 10
done
echo '"after decompress, ls $zst_file*' >> ${WORK_DIR}/custom-out.txt
ls -l ${full_dir} >> ${WORK_DIR}/custom-out.txt

# get the filename without extension
# arr=($(echo $f_dir | tr "/" "\n"))
# fname=${arr[2]}
# bname=${f_dir%.*}

bname=$m

# 3. run the programs to dedisperse and search
echo "DIR_DEDISP:$DIR_DEDISP/$bname"
mkdir -p $DIR_DEDISP/$bname
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In mkdir:$bname, ret-code:$code" >&2 && exit 11

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

cd $DIR_DEDISP/$bname
rfifind $RFIARGS -o RFIfile ${full_dir}/*.fits
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In rfi-find:$f_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname && exit 12

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

# compress the files again

for full_name in $( ls ${full_dir}/*.fits )
do
    # full_name="$DIR_FITS/${f_dir}"
    zst_file=${full_name}.zst
    echo "full_name:${full_name}" >> ${WORK_DIR}/custom-out.txt
    [ -f "${full_name}" ] && zstd --rm -f ${full_name}

    [[ ! -f $zst_file ]] && echo "[ERROR] In checking file exits:$zst_file, ret-code:$code" >&2 && exit 13
done

echo $DIR_FITS/${m} >> ${WORK_DIR}/input-files.txt
echo $DIR_DEDISP/$bname/RFIfile_rfifind.mask >> ${WORK_DIR}/output-files.txt

echo "send message to sink job"
echo ${m} > ${WORK_DIR}/messages.txt

exit $code