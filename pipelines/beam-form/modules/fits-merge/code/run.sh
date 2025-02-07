#!/bin/bash

source functions.sh

# DOWN_SAMPLER_ENABLED

if [ $LOCAL_INPUT_ROOT ]; then
    DIR_1CHZ=$(get_host_path "${LOCAL_INPUT_ROOT}/mwa/1chz")
else
    DIR_1CHZ=/cluster_data_root/mwa/1chz
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_24CH=$(get_host_path "${LOCAL_OUTPUT_ROOT}/mwa/24ch")
else
    DIR_24CH=/cluster_data_root/mwa/24ch
fi
echo "DIR_1CHZ:${DIR_1CHZ}, DIR_24CH:${DIR_24CH}"

# 应该是 ${单通道目录根}/${观测号}/指向号/${起始时间戳}_${结尾时间戳}
# m="1257010784/p00001/t1257010986_1257011185"
cd ${DIR_1CHZ}/$1
zstd -d --rm *.zst

input_files=$(ls *.fits)
echo input_files:${input_files} >> ${WORK_DIR}/custom-out.txt
splice_psrfits ${input_files} ${WORK_DIR}/all; code=$?
[[ $code -ne 0 ]] && echo "[ERROR]exit after splice_psrfits, error_code:$code"  >> ${WORK_DIR}/custom-out.txt && exit $code

# Swap the time_range and the pointing parts
IFS='/' read -r dataset pointing time_range <<< $(echo "$1")
new_id="${dataset}/${pointing}/${time_range}"

output_file=${DIR_24CH}/$new_id.fits
output_dir=$(dirname ${output_file})
filename=$(basename ${output_file})

echo "new feature for local-copy" >> ${WORK_DIR}/custom-out.txt

cd ${WORK_DIR} && mv -f all*.fits ${filename} 
# mkdir -p $(dirname ${output_file}) && mv -f ${WORK_DIR}/all*.fits ${output_file}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] rename fits file "  >> ${WORK_DIR}/custom-out.txt && exit $code

mkdir -p ${output_dir} && cd ${output_dir} && zstd -f --rm ${WORK_DIR}/${filename} -o ${filename}.zst
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] zstd compress target fits file "  >> ${WORK_DIR}/custom-out.txt && exit $code

echo "${output_file}.zst" > ${WORK_DIR}/output-files.txt
[ "$KEEP_TARGET_FILE" = "no" ] && echo "${output_file}.zst" >> ${WORK_DIR}/removed-files.txt

full_path="${DIR_1CHZ}/$1"
echo [DEBUG]full_path:$full_path
[ "$KEEP_SOURCE_FILE" = "no" ] && echo $full_path >> ${WORK_DIR}/removed-files.txt
echo $full_path >> ${WORK_DIR}/input-files.txt

echo "send-message to sink-job"
echo $new_id > ${WORK_DIR}/messages.txt

exit $code
