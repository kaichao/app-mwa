#!/bin/bash

# command line args:
# $m: directory of the input files. the input files are in the $DIR_DEDISP/$m.tar.zst file.

# environment variables:
# $SEARCHARGS           arguments for accelsearch_gpu_4

# 1. set the input / output / medium file directory

# m="/1257010784/00017/1"
# source /root/.bashrc
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
if [ $LOCAL_INPUT_ROOT ]; then
    DIR_DEDISP="/local${LOCAL_OUTPUT_ROOT}/mwa/dedisp"
else
    DIR_DEDISP=/data/mwa/dedisp
fi
if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_PNG="/local${LOCAL_OUTPUT_ROOT}/mwa/png"
else
    DIR_PNG=/data/mwa/png
fi

m=$1
# m=${bname}/${linenum}
# parse m
bname=$(dirname $m)
linenum=$(basename $m)
full_dir="$DIR_DEDISP/${m}"

echo "DIR_DEDISP:$DIR_DEDISP/$bname"
echo "DIR_PNG:$DIR_PNG/$bname"

# 2. check if the input file ($DIR_DEDISP/$bname.tar.zst) exists
[[ ! -f "$DIR_DEDISP/$m.tar.zst" ]] && echo "[ERROR] In checking file exits:$DIR_DEDISP/$m.tar.zst" >&2 && exit 10

# 3. untar the input file. remove the input file.
cd $DIR_DEDISP/$bname
zstd -d --rm $linenum.tar.zst && tar -xf $linenum.tar && rm -f $linenum.tar
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In untar:$DIR_DEDISP/$bname.tar.zst, ret-code:$code" >&2 && exit 11

mkdir -p $DIR_PNG/$bname/$linenum
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In mkdir:$bname, ret-code:$code" >&2 && exit 12
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

cd $DIR_DEDISP/$bname/$linenum
# 4. now run accelsearch_gpu_4. we will later modify the code to run accelsearch_gpu_4 single time for all the files.
# for filename in $( ls *.dat )
# do
#     datname=$(basename $filename .dat)
#     realfft $filename
#     accelsearch_gpu_4 -cuda 0 $SEARCHARGS $datname.fft | grep Total
#     rm $datname.fft
# done
realfft *.dat && ls *.fft | xargs -n 1 accelsearch_gpu_4 -cuda 0 $SEARCHARGS | grep Total
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt && du . -sh
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In accelsearh:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname/$linenum && exit 13

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
# 5. run ACCEL_sift.py
python3 /code/presto/examplescripts/ACCEL_sift.py > candidates.txt
[[ $code -ne 0 ]] && echo "[ERROR]Error in ACCEL_sift:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname/$linenum && exit 14

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

# 6. parse candidates.txt, fold .dat file at each dm
/app/bin/fold_dat.py $full_dir candidates.txt
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]Error in folding:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname/$linenum && exit 15

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

echo "DIR_DEDISP:$DIR_DEDISP/$bname/$linenum"
echo "DIR_PNG:$DIR_PNG/$bname/$linenum"

# copy the result to target dir

mv *.pfd* $DIR_PNG/$bname/$linenum && mv candidates.txt $DIR_PNG/$bname/$linenum
code=$?

# record input and output files
echo $DIR_PNG/$bname/$linenum >> ${WORK_DIR}/output-files.txt
echo $DIR_FITS/$bname/$linenum.tar.zst >> ${WORK_DIR}/input-files.txt

# clean up
rm -rf $DIR_DEDISP/$bname/$linenum
# [ "$KEEP_SOURCE_FILE" == "no" ] && rm -f $DIR_FITS/$f_dir
# [[ $code -eq 0 ]] && [ "$KEEP_SOURCE_FILE" == "no" ] && echo $DIR_FITS/$f_dir >> ${WORK_DIR}/removed-files.txt

exit $code
