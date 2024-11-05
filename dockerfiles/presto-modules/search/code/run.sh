#!/bin/bash

# command line args:
# $m: directory of the input files. the input files are in the $DIR_DEDISP/$m.tar.zst file.

# environment variables:
# $SEARCHARGS           arguments for accelsearch_gpu_4

# 1. set the input / output / medium file directory

# m="/1257010784/00017/dm1/group1.tar.zst"
# source /root/.bashrc
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
if [ $LOCAL_INPUT_ROOT ]; then
    DIR_DEDISP="/local${LOCAL_INPUT_ROOT}/mwa/dedisp"
else
    DIR_DEDISP=/cluster_data_root/mwa/dedisp
fi
if [ $LOCAL_INPUT_ROOT ]; then
    DIR_TAR="/local${LOCAL_INPUT_ROOT}/mwa/dedisp/tar"
else
    DIR_DEDISP=/cluster_data_root/mwa/dedisp/tar
fi

m=$1
# m=${bname}/${dm_group}
# parse m
bname=$(dirname $m)
filenm=$(basename $m)
dm_group=${filenm%%.*}
full_dir="$DIR_DEDISP/${m}"

echo "DIR_DEDISP:$DIR_DEDISP/$bname"
echo "DIR_TAR:$DIR_TAR/$bname"

# 2. check if the input file ($DIR_DEDISP/$m) exists
[[ ! -f "$DIR_TAR/$m" ]] && echo "[ERROR] In checking file exits:$DIR_TAR/$m" >&2 && exit 10

# 3. untar the input file. remove the input file.
[[ ! -d "$DIR_DEDISP/bname" ]] && mkdir -p $DIR_DEDISP/$bname
cd $DIR_DEDISP/$bname
mv $DIR_TAR/$m ./
zstd -d --rm $dm_group.tar.zst && tar -xf $dm_group.tar && rm -f $dm_group.tar
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In untar:$DIR_DEDISP/$bname.tar.zst, ret-code:$code" >&2 && exit 11
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

cd $DIR_DEDISP/$bname/$dm_group
# 4. now run accelsearch_gpu_4. we will later modify the code to run accelsearch_gpu_4 single time for all the files.
# for filename in $( ls *.dat )
# do
#     datname=$(basename $filename .dat)
#     realfft $filename
#     accelsearch_gpu_4 -cuda 0 $SEARCHARGS $datname.fft | grep Total
#     rm $datname.fft
# done
realfft *.dat && ls *.fft && accelsearch_gpu_multifile -cuda 0 $SEARCHARGS *.fft | grep Total
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] In accelsearh:$full_dir, ret-code:$code" >&2 && rm -rf $DIR_DEDISP/$bname/$dm_group && exit 13
rm *.fft
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
echo $DIR_TAR/$bname/$dm_group.tar.zst >> ${WORK_DIR}/input-files.txt
echo $DIR_DEDISP/$bname/$dm_group >> ${WORK_DIR}/output-files.txt

# send message to sink job
echo ${bname}/${dm_group} >> ${WORK_DIR}/messages.txt

exit $code
