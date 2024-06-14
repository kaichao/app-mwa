#!/bin/bash

source $(dirname $0)/functions.sh

# usage: downsample all .fits file in the input directory,
# and save the result to the output directory.

# downsample command:

# psrfits_subband -dstime 4 -o 1165080856_J0630_4dt 1165080856_J0630_5s.fits 
##### only downsamples in time

# psrfits_subband -dstime 4 -outbits 4 -adjustlevels -o 1165080856_J0630_4dt 1165080856_J0630_5s.fits
##### also changes the output file to 4-bit format


# m="1257010784/1257010786_1257010795/00001/ch123.fits"

# 1. set the input / output directory
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

# remove existing intermediate files in output dir
rm -f ${DIR_1CHX}/${m}*

if [ -f "${DIR_1CH}/${m}.zst" ]; then
    # 解压文件并删除原始文件
    zstd -d --rm "${DIR_1CH}/${m}.zst"
fi

echo [DEBUG]input file:
ls -l ${DIR_1CH}/${m}
echo [DEBUG]

# 3. run the programs to downsample the files
psrfits_subband -dstime ${DOWNSAMP_FACTOR_TIME} -o ${DIR_1CHX}/${m} ${DIR_1CH}/${m}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] psrfits_subband " >&2 && exit $code

# [ "$KEEP_SOURCE_FILE" == "no" ] && rm -f ${DIR_1CH}/${m}
[ "$KEEP_SOURCE_FILE" == "no" ] && echo "${DIR_1CH}/${m}" > ${WORK_DIR}/removed-files.txt

# rename file to normalized
mv ${DIR_1CHX}/${m}_0001.fits ${DIR_1CHX}/${m} && zstd --long -T8 --rm ${DIR_1CHX}/${m}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] rename fits file and zstd compress " >&2 && exit $code

# 检查输入、输出文件的大小比例是否合理？
post_check "${DIR_1CH}/${m}" "${DIR_1CHX}/${m}.zst"
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] post_check ${m} " >> ${WORK_DIR}/custom-out.txt && exit $code

echo "${DIR_1CH}/${m}" > ${WORK_DIR}/input-files.txt
echo "${DIR_1CHX}/${m}.zst" > ${WORK_DIR}/output-files.txt
echo "${m}.zst" >> ${WORK_DIR}/messages.txt

exit $code
