#!/bin/bash

# command line args:
# $m: file name to be executed

# environment variables:
# $NSUB                 nsub for prepsubband_gpu
# $RFIARGS              arguments for rfifind
# $SEARCHARGS           arguments for accelsearch_gpu_4

# 1. set the input / output / medium file directory

# m="/1257010784/1257010786_1257011025/00024.fits"
source /root/.bashrc
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
if [ $LOCAL_INPUT_ROOT ]; then
    DIR_FITS="/local${LOCAL_INPUT_ROOT}/mwa/24ch"
else
    DIR_FITS=/data/mwa/24ch
fi
if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_PNG="/local${LOCAL_OUTPUT_ROOT}/mwa/png"
else
    DIR_PNG=/data/mwa/png
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_DEDISP="/local${LOCAL_OUTPUT_ROOT}/mwa/dedisp"
else
    DIR_DEDISP=/data/mwa/dedisp
fi

# decompress zst file
m=$1
f_dir=${m}.fits

full_name="$DIR_FITS/${f_dir}"
zst_file="${full_name}.zst"
echo "full_name:${full_name}" >> ${WORK_DIR}/custom-out.txt
echo '"before decompress, ls $zst_file*' >> ${WORK_DIR}/custom-out.txt
ls -l $(dirname ${zst_file}) >> ${WORK_DIR}/custom-out.txt

[ -f "${zst_file}" ] && cd $(dirname ${zst_file}) && zstd -d --rm -f $(basename ${zst_file})

# cd $DIR_FITS/$(dirname $1) && [ -f "$(basename $1).fits.zst" ] && zstd -d --rm -f $(basename $1).fits.zst
echo '"after decompress, ls $zst_file*' >> ${WORK_DIR}/custom-out.txt
ls -l $(dirname ${zst_file}) >> ${WORK_DIR}/custom-out.txt

# 2. check if the file exists

# readfile $DIR_FITS/$f_dir
# code=$?
# [[ $code -ne 0 ]] && echo "[ERROR]Error in checking file exits:$fdir, ret-code:$code" >&2 && exit 10
[[ ! -f $DIR_FITS/$f_dir ]] && echo "[ERROR] In checking file exits:$f_dir, ret-code:$code" >&2 && exit 10

# get the filename without extension
# arr=($(echo $f_dir | tr "/" "\n"))
# fname=${arr[2]}
# bname=${f_dir%.*}

bname=$1

# 3. run the programs to dedisperse and search
echo "DIR_DEDISP:$DIR_DEDISP/$bname"
echo "DIR_PNG:$DIR_PNG/$bname"
mkdir -p $DIR_DEDISP/$bname $DIR_PNG/$bname
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In mkdir:$bname, ret-code:$code" >&2 && exit 11

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

cd $DIR_DEDISP/$bname
rfifind $RFIARGS -o RFIfile $DIR_FITS/$f_dir
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]Error in dedispersion:$f_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname && exit 12

echo 2222222

for i in {1..9}
do
    export LINENUM=$i
    /app/bin/dedisp_all.py $DIR_FITS/$f_dir RFIfile_rfifind.mask
    code=$?
    [[ $code -ne 0 ]] && echo "[ERROR]Error in dedispersion:$f_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname && exit 13

    date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

    python3 /code/presto/examplescripts/ACCEL_sift.py > candidates$i.txt
    [[ $code -ne 0 ]] && echo "[ERROR]Error in ACCEL_sift:$f_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname && exit 14

    date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

    # 4. parse candidates.txt, fold at each dm
    /app/bin/fold_dat.py $DIR_FITS/$f_dir candidates$i.txt
    code=$?
    [[ $code -ne 0 ]] && echo "[ERROR]Error in folding:$f_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname && exit 15

    date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
    
    # copy the result to target dir

    mv *.pfd* $DIR_PNG/$bname && mv candidates* $DIR_PNG/$bname
    code=$?
    rm *DM*

done
# clean up
rm -r $DIR_DEDISP/$bname
exit $code
