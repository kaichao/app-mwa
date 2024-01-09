#!/bin/bash

# command line args:
# $m: file name to be executed

# environment variables:
# $NSUB                 nsub for prepsubband_gpu
# $RFIARGS              arguments for rfifind
# $SEARCHARGS           arguments for accelsearch_gpu_4


# 1. set the input / output / medium file directory

# m="mwa/24ch~/1257010784/1257010786_1257011025/00024.fits"
source /root/.bashrc

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
DIR_MID=/data/mwa/dedisp

# 2. check if the file exists
m=$1
arr=($(echo $m | tr "~" "\n"))
f_dir=${arr[1]}
if [ ! -f "$DIR_FITS/$f_dir" ]; then
    echo "[ERROR]invalid input message:$f_dir" >&2 && exit 5
fi
# get the filename without extension
arr=($(echo $f_dir | tr "/" "\n"))
# fname=${arr[2]}
bname=${f_dir%.*}
# 3. run the programs to dedisperse and search

mkdir -p $DIR_MID/$bname
cd $DIR_MID/$bname
rfifind $RFIARGS -o RFIfile $DIR_FITS/$f_dir

/app/bin/dedisp.py $DIR_FITS/$f_dir RFIfile_rfifind.mask

code=$?
if [ $code -ne 0 ]; then 
    echo "[ERROR]Error in dedispersion:$f_dir" >&2
    rm ${bname}_*
    cd /work
    exit $code
fi

python3 /code/presto/examplescripts/ACCEL_sift.py > candidates.txt
source /root/.bashrc
env
ls /app/bin
# 4. parse candidates.txt, fold at each dm
/app/bin/fold.py $DIR_FITS/$f_dir candidates.txt
code=$?
if [ $code -ne 0 ]; then 
    echo "[ERROR]Error in folding:$f_dir" >&2
    rm ${bname}_*
    cd /work
    exit $code
fi

# copy the result to target dir
mkdir -p $DIR_PNG/$bname
mv *.pfd* $DIR_PNG/$bname
mv candidates.txt $DIR_PNG/$bname

# clean up
rm ./candidates.txt ${bname}_*

code=$?
cd /work
exit $code