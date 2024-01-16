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

# m="1257010784/1257010786_1257010795/00001/ch123.fits"

if [ $LOCAL_INPUT_ROOT ]; then
    DIR_1CH="/local${LOCAL_INPUT_ROOT}/mwa/1ch"
else
    DIR_1CH=/data/mwa/1ch
fi
if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_1CHX="/local${LOCAL_OUTPUT_ROOT}/mwa/1chx"
else
    DIR_1CHX=/data/mwa/1chx
fi
echo "DIR_1CH:${DIR_1CH}, DIR_1CHX:${DIR_1CHX}"

# 2. check if the directory exists
m=$1
dir=$(dirname $DIR_1CHX/$m)
mkdir -p $dir; code=$?
[[ $code -ne 0 ]] && echo "[ERROR] mkdir $dir" >&2 && exit $code

# if [ ! -f "$DIR_FITS/$m" ]; then
#     echo "[ERROR]invalid input message:$f_dir" >&2 && exit 5
# fi

# 3. run the programs to downsample the files

psrfits_subband -dstime ${DOWNSAMP_FACTOR_TIME} -o ${DIR_1CHX}/${m} ${DIR_1CH}/${m}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] psrfits_subband " >&2 && exit $code

[ "$KEEP_SOURCE_FILE" == "no" ] && rm -f ${DIR_1CH}/${m}

# rename file to normalized
mv ${DIR_1CHX}/${m}_0001.fits ${DIR_1CHX}/${m} && zstd --rm ${DIR_1CHX}/${m}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] rename fits file and zstd compress " >&2 && exit $code

echo "${m}.zst" >> /work/messages.txt

exit $code
