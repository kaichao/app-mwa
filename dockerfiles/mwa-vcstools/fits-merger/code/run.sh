#!/bin/bash

# DOWN_SAMPLER_ENABLED

if [ $LOCAL_INPUT_ROOT ]; then
    DIR_1CHX="/local${LOCAL_INPUT_ROOT}/mwa/1chx"
else
    DIR_1CHX=/data/mwa/1chx
fi
if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_24CH="/local${LOCAL_OUTPUT_ROOT}/mwa/24ch"
else
    DIR_24CH=/data/mwa/24ch
fi
echo "DIR_1CHX:${DIR_1CHX}, DIR_24CH:${DIR_24CH}"

# 应该是 ${单通道目录根}/${观测号}/指向号/${起始时间戳}_${结尾时间戳}
# m="1257010784/p00001/t1257010986_1257011185"
cd ${DIR_1CHX}/$1
zstd -d --rm *.zst

input_files=$(ls *.fits)
echo input_files:${input_files}
splice_psrfits ${input_files} ${WORK_DIR}/all; code=$?
[[ $code -ne 0 ]] && echo exit after splice_psrfits, error_code:$code >&2 && exit $code

# Swap the time_range and the pointing parts
IFS='/' read -r dataset pointing time_range <<< $(echo "$1")
new_id="${dataset}/${pointing}/${time_range}"

output_file=${DIR_24CH}/$new_id.fits

mkdir -p $(dirname ${output_file}) && mv -f ${WORK_DIR}/all*.fits ${output_file}
code=$?
[[ $code -ne 0 ]] && echo "mv fits file to target dir" >&2 && exit $code

dir=$(dirname ${output_file})
file=$(basename ${output_file})
cd ${dir} && rm -f ${file}.zst && zstd --rm ${file}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] zstd compress target fits file " >&2 && exit $code

echo "${output_file}.zst" > ${WORK_DIR}/output-files.txt
[ "$KEEP_TARGET_FILE" = "no" ] && echo "${output_file}.zst" >> ${WORK_DIR}/removed-files.txt

full_path="${DIR_1CHX}/$1"
echo [DEBUG]full_path:$full_path
[ "$KEEP_SOURCE_FILE" = "no" ] && echo $full_path >> ${WORK_DIR}/removed-files.txt
echo $full_path >> ${WORK_DIR}/input-files.txt

echo "send-message to sink-job"
echo $new_id > ${WORK_DIR}/messages.txt

exit $code
