#!/bin/bash

source functions.sh

# DOWN_SAMPLER_ENABLED
# env >> ${WORK_DIR}/auxout.txt

input_root=$(get_header "$2" "input_root")
if [ $input_root ]; then
    DIR_1CHY=$(get_host_path "${input_root}/mydata/mwa/1chy")
else
    DIR_1CHY=/cluster_data_root/mwa/1chy
fi

output_root=$(get_header "$2" "output_root")
if [ $output_root ]; then
    DIR_24CH=$(get_host_path "${output_root}/mwa/24ch")
else
    DIR_24CH=/cluster_data_root/mwa/24ch
fi
echo "DIR_1CHY:${DIR_1CHY}, DIR_24CH:${DIR_24CH}" >> ${WORK_DIR}/auxout.txt
echo "work_sub_dir:${DIR_1CHY}/$1" >> ${WORK_DIR}/auxout.txt

# 应该是 ${单通道目录根}/${观测号}/指向号/${起始时间戳}_${结尾时间戳}
# m="1257010784/p00001/t1257010986_1257011185"

# Check if directory exists
if [ ! -d "${DIR_1CHY}/$1" ]; then
    echo "[ERROR] Directory ${DIR_1CHY}/$1 does not exist" >> ${WORK_DIR}/auxout.txt
    exit 101
fi

cd "${DIR_1CHY}/$1"
pwd >> ${WORK_DIR}/auxout.txt
ls -l >> ${WORK_DIR}/auxout.txt

# Only decompress if zst files exist
if ls *.zst 1> /dev/null 2>&1; then
    ls -l *.zst >> ${WORK_DIR}/auxout.txt
    zstd -d --rm *.zst
fi

# Check if fits files exist
if ! ls *.fits 1> /dev/null 2>&1; then
    echo "[ERROR] No fits files found in directory" >> ${WORK_DIR}/auxout.txt
    exit 102
fi

input_files=$(ls *.fits)
echo input_files:${input_files} >> ${WORK_DIR}/auxout.txt
splice_psrfits ${input_files} ${WORK_DIR}/all; code=$?
[[ $code -ne 0 ]] && echo "[ERROR]exit after splice_psrfits, error_code:$code"  >> ${WORK_DIR}/auxout.txt && exit $code

# Swap the time_range and the pointing parts
IFS='/' read -r dataset pointing time_range <<< $(echo "$1")
new_id="${dataset}/${pointing}/${time_range}"

output_file=${DIR_24CH}/$new_id.fits
output_dir=$(dirname ${output_file})
filename=$(basename ${output_file})

echo "new feature for local-copy" >> ${WORK_DIR}/auxout.txt
echo "output_dir:$output_dir" >> ${WORK_DIR}/auxout.txt

cd ${WORK_DIR} && mv -f all*.fits ${filename} 
# mkdir -p $(dirname ${output_file}) && mv -f ${WORK_DIR}/all*.fits ${output_file}
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] rename fits file "  >> ${WORK_DIR}/auxout.txt && exit $code

bw_limit=$(get_header "$2" "bw_limit")
# BW_LIMIT  "500k"/"1m"
if [ -n "$bw_limit" ]; then
    if [ "$ZSTD_TARGET_FILE" = "no" ]; then
        cmd="cat ${WORK_DIR}/${filename} | pv -q -L $bw_limit > ${filename}; rm -f ${WORK_DIR}/${filename}"
    else
        cmd="zstd -c --rm ${WORK_DIR}/${filename} | pv -q -L $bw_limit > ${filename}.zst"
    fi
else
    if [ "$ZSTD_TARGET_FILE" = "no" ]; then
        cmd="mv -f ${WORK_DIR}/${filename} ."
    else
        cmd="zstd -f --rm ${WORK_DIR}/${filename} -o ${filename}.zst"
    fi
fi
mkdir -p "${output_dir}" && cd "${output_dir}" && eval $cmd
# mkdir -p ${output_dir} && cd ${output_dir} && zstd -f --rm ${WORK_DIR}/${filename} -o ${filename}.zst
code=$?
[[ $code -ne 0 ]] && echo "[ERROR] zstd compress target fits file "  >> ${WORK_DIR}/auxout.txt && exit $code

echo "${output_file}.zst" > ${WORK_DIR}/output-files.txt
[ "$KEEP_TARGET_FILE" = "no" ] && echo "${output_file}.zst" >> ${WORK_DIR}/removed-files.txt

full_path="${DIR_1CHY}/$1"
echo [DEBUG]full_path:$full_path
[ "$KEEP_SOURCE_FILE" = "no" ] && echo $full_path >> ${WORK_DIR}/removed-files.txt
echo $full_path >> ${WORK_DIR}/input-files.txt

echo "send-message to sink-module"
echo $new_id > ${WORK_DIR}/sink-tasks.txt

exit $code
