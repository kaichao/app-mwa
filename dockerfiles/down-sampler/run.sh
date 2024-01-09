#!/bin/bash

# usage: downsample all .fits file in the input directory,
# and save the result to the output directory.

# downsample command:

# psrfits_subband -dstime 4 -o 1165080856_J0630_4dt 1165080856_J0630_5s.fits 
##### only downsamples in time

# psrfits_subband -dstime 4 -outbits 4 -adjustlevels -o 1165080856_J0630_4dt 1165080856_J0630_5s.fits
##### also changes the output file to 4-bit format

# command line args:
# $1: the input message


# 1. set the input / output directory

# m="/directory/root~/path/to/downsample"

if [ $LOCAL_INPUT_ROOT ]; then
    DIR_IN="/local${LOCAL_INPUT_ROOT}/mwa/in"
else
    DIR_IN=/data/mwa/in
fi
if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_OUT="/local${LOCAL_OUTPUT_ROOT}/mwa/out"
else
    DIR_OUT=/data/mwa/out
fi


# 2. check if the directory exists
m=$1
arr=($(echo $m | tr "~" "\n"))
f_dir=${arr[1]}
if [ ! -d "$DIR_IN/$f_dir" ]; then
    echo "[ERROR]invalid input message:$f_dir" >&2 && exit 5
fi
# 3. run the programs to downsample the files

for file in $(ls ${DIR_IN}/${f_dir}/*.fits)
do
    filename=$(basename $file)
    filename=${filename%.*}
    psrfits_subband -dstime ${DOWNSAMP_FACTOR_TIME} -o ${DIR_OUT}/${f_dir}/${filename}_4dt ${DIR_IN}/${f_dir}/${filename}.fits
done

code=$?
exit $code
