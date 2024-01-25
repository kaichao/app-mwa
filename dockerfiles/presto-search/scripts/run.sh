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

# 2. check if the file exists
m=$1
f_dir=${m}
if [ ! -f "$DIR_FITS/$f_dir" ]; then
    echo "[ERROR]invalid input message:$f_dir" >&2 && exit 5
fi
# get the filename without extension
# arr=($(echo $f_dir | tr "/" "\n"))
# fname=${arr[2]}
bname=${f_dir%.*}
# 3. run the programs to dedisperse and search

mkdir -p $DIR_DEDISP/$bname
mkdir -p $DIR_PNG/$bname

code=$?
if [ $code -ne 0 ]; then 
    echo "[ERROR]Error in mkdir:$bname" >&2
    exit $code
fi

cd $DIR_DEDISP/$bname
rfifind $RFIARGS -o RFIfile $DIR_FITS/$f_dir
code=$?
if [ $code -ne 0 ]; then 
    echo "[ERROR]Error in dedispersion:$f_dir" >&2
    cd /work
    rm -r $DIR_DEDISP/$bname
    exit $code
fi

/app/bin/dedisp.py $DIR_FITS/$f_dir RFIfile_rfifind.mask

code=$?
if [ $code -ne 0 ]; then 
    echo "[ERROR]Error in dedispersion:$f_dir" >&2
    cd /work
    rm -r $DIR_DEDISP/$bname
    exit $code
fi

python3 /code/presto/examplescripts/ACCEL_sift.py > candidates.txt
if [ $code -ne 0 ]; then 
    echo "[ERROR]Error in ACCEL_sift:$f_dir" >&2
    exit $code
fi
# 4. parse candidates.txt, fold at each dm
/app/bin/fold.py $DIR_FITS/$f_dir candidates.txt
code=$?
if [ $code -ne 0 ]; then 
    echo "[ERROR]Error in folding:$f_dir" >&2
    cd /work
    rm -r $DIR_DEDISP/$bname
    exit $code
fi

# copy the result to target dir

mv *.pfd* $DIR_PNG/$bname
mv candidates.txt $DIR_PNG/$bname

# clean up
code=$?
cd /work
rm -r $DIR_DEDISP/$bname
exit $code