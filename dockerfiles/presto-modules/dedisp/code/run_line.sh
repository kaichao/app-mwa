#!/bin/bash

# command line args:
# $m: file name to be executed

# environment variables:
# $NSUB                 nsub for prepsubband_gpu
# $RFIARGS              arguments for rfifind
# $SEARCHARGS           arguments for accelsearch_gpu_4

# 1. set the input / output / medium file directory

# m="/1257010784/00017/1"
source /root/.bashrc
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

m0=$1
m=${m0%/*}
export LINENUM=${m0##*/}
echo "DIR_FITS:$DIR_FITS/$m"
# f_dir=${m}.fits
full_dir="$DIR_FITS/${m}"
bname=$m

# the file have already been uncompressed.
# for zst_file in $( ls ${full_dir}/*.zst )
# do
#     # full_name="$DIR_FITS/${f_dir}"
#     full_name=${zst_file%.zst}
#     echo "full_name:${full_name}" >> ${WORK_DIR}/custom-out.txt
#     [ -f "${zst_file}" ] && zstd -d --rm -f ${zst_file}

#     # cd $DIR_FITS/$(dirname $1) && [ -f "$(basename $1).fits.zst" ] && zstd -d --rm -f $(basename $1).fits.zst
#     # 2. check if the file exists

#     # readfile $DIR_FITS/$f_dir
#     # code=$?
#     # [[ $code -ne 0 ]] && echo "[ERROR]Error in checking file exits:$fdir, ret-code:$code" >&2 && exit 10
#     [[ ! -f $full_name ]] && echo "[ERROR] In checking file exits:$full_name, ret-code:$code" >&2 && exit 10
# done

# 3. run the programs to dedisperse and search
echo "DIR_DEDISP:$DIR_DEDISP/$bname"
cd $DIR_DEDISP/$bname
[[ ! -f "RFIfile_rfifind.mask" ]] && echo "[ERROR] In checking file exits:RFIfile_rfifind.mask, ret-code:$code" >&2 && exit 10

echo "LINENUM:${LINENUM}"
mkdir -p ${DIR_DEDISP}/${bname}/dm${LINENUM}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In mkdir:$bname, ret-code:$code" >&2 && exit 11


cd dm${LINENUM}
/app/bin/dedisp_line.py $full_dir ../RFIfile_rfifind.mask
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In dedispersion:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname/dm${LINENUM} && exit 13

cd ..
du -sh
tar -cf dm${LINENUM}.tar ./dm${LINENUM} && rm -rf ./dm${LINENUM}
zstd --rm -f dm${LINENUM}.tar
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

echo $DIR_FITS/${m} >> ${WORK_DIR}/input-files.txt
echo $DIR_DEDISP/$bname/dm${LINENUM}.tar.zst >> ${WORK_DIR}/output-files.txt

echo "send message to sink job"
echo ${bname}/dm${LINENUM} > ${WORK_DIR}/messages.txt
exit $code