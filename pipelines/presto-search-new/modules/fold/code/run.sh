#!/bin/bash

# command line args:
# $m: directory of the input files. the input files are in the $DIR_DEDISP/$m.tar.zst file.

# environment variables:
# 1. set the input / output file directory

# m="1257010784/00017/dm1"
# source /root/.bashrc
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
if [ $LOCAL_INPUT_ROOT ]; then
    DIR_DEDISP="/local_data_root${LOCAL_INPUT_ROOT}/mwa/dedisp"
else
    DIR_DEDISP=/cluster_data_root/mwa/dedisp
fi
if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_PNG="/local_data_root${LOCAL_OUTPUT_ROOT}/mwa/png"
else
    DIR_PNG=/cluster_data_root/mwa/png
fi

m=$1
# m=${bname}/${dm_group}
# parse m
bname=$(dirname $m)
dm_group=$(basename $m)
full_dir="$DIR_DEDISP/${m}"

echo "DIR_DEDISP:$DIR_DEDISP/$bname"
echo "DIR_PNG:$DIR_PNG/$bname"

# 3. move all file from sub-directories to current directory
cd $DIR_DEDISP/$m
for subdir in $( ls | grep group )
do
    echo $subdir
    mv ./${subdir}/* ./
done
# 5. run ACCEL_sift.py
python3 /code/presto/examplescripts/ACCEL_sift.py > candidates.txt
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]Error in ACCEL_sift:$full_dir, ret-code:$code" >&2 && ls >> ${WORK_DIR}/output-files.txt && rm -rf $DIR_DEDISP/$bname/$dm_group && exit 14

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

# 6. parse candidates.txt, fold .dat file at each dm
/app/bin/fold_dat.py $full_dir candidates.txt
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]Error in folding:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname/$dm_group && exit 15
cat candidates.txt
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

echo "DIR_DEDISP:$DIR_DEDISP/$bname/$dm_group"
echo "DIR_PNG:$DIR_PNG/$bname/$dm_group"

# copy the result to target dir
mkdir -p $DIR_PNG/$bname/$dm_group
mv *.pfd* $DIR_PNG/$bname/$dm_group && mv candidates.txt $DIR_PNG/$bname/$dm_group && mv *.cand $DIR_PNG/$bname/$dm_group
cd $DIR_PNG/$bname && tar -cf ${dm_group}.tar ./${dm_group} && rm -rf ./${dm_group}
zstd --rm -f ${dm_group}.tar
code=$?

# record input and output files
echo $DIR_PNG/$bname/$dm_group.tar.zst >> ${WORK_DIR}/output-files.txt

echo $DIR_DEDISP/$bname/$dm_group >> ${WORK_DIR}/input-files.txt

# send messages to sink job
echo "send message to sink job"
echo $m.tar.zst > ${WORK_DIR}/messages.txt

# clean up
rm -rf $DIR_DEDISP/$bname/$dm_group

exit $code
