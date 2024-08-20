#!/bin/bash

# command line args:
# $m: file name to be executed

# environment variables:
# $NSUB                 nsub for prepsubband_gpu
# $SEARCHARGS           arguments for accelsearch_gpu_4

# 1. set the input / output / medium file directory

# m="/1257010784/00017/1"
source /root/.bashrc
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

m0=$1
m=${m0%/*}
export GRPNUM=${m0##*/}
echo "DIR_FITS:$DIR_FITS/$m"
# f_dir=${m}.fits
full_dir="$DIR_FITS/${m}"
bname=$m

# the file have already been uncompressed.
# 3. run the programs to dedisperse and search
echo "DIR_DEDISP:$DIR_DEDISP/$bname"
cd $DIR_DEDISP/$bname
[[ ! -f "RFIfile_rfifind.mask" ]] && echo "[ERROR] In checking file exits:RFIfile_rfifind.mask, ret-code:$code" >&2 && exit 10

echo "GRPNUM:${GRPNUM}"
mkdir -p ${DIR_DEDISP}/${bname}/group${GRPNUM}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In mkdir:$bname, ret-code:$code" >&2 && exit 11

cd group${GRPNUM}
/app/bin/dedisp_line_new.py $full_dir ../RFIfile_rfifind.mask
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In dedispersion:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname/group${GRPNUM} && exit 13
LINENUM=$( cat ./linenum.txt ) && echo "LINENUM = ${LINENUM}"
rm ./linenum.txt
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
# next, run accelsearch on these data.
realfft *.dat
accelsearch_gpu_multifile -cuda 0 -ncpus $NCPUS $SEARCHARGS *.fft | grep Total
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In accelsearch:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname/group${GRPNUM} && exit 14
rm *.fft
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
# move all the files into ../$linenum
cd ..
# du -sh
# tar -cf group${GRPNUM}.tar ./group${GRPNUM} && rm -rf ./group${GRPNUM}
# zstd --rm -f group${GRPNUM}.tar
[[ ! -d dm${LINENUM} ]] && mkdir -p dm${LINENUM}
mv ./group${GRPNUM} ./dm${LINENUM}

echo $DIR_FITS/${m} >> ${WORK_DIR}/input-files.txt
echo $DIR_DEDISP/$bname/dm${LINENUM}/group${GRPNUM} >> ${WORK_DIR}/output-files.txt

echo "send message to sink job"
echo ${bname}/dm${LINENUM}/group${GRPNUM} > ${WORK_DIR}/messages.txt
exit $code