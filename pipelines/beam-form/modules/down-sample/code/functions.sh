#!/bin/bash

# 检查输入文件/输出文件压缩比是否在合理范围？
function post_check() {
    input_file=$1
    output_file=$2

    # 检查input_file是否存在
    if [ ! -f "$input_file" ]; then
        echo "[Error] Input file $input_file does not exist." >> ${WORK_DIR}/custom-out.txt
        return 91
    fi

    # 检查output_file是否存在
    if [ ! -f "$output_file" ]; then
        echo "[Error] Output file $output_file does not exist." >> ${WORK_DIR}/custom-out.txt
        return 92
    fi

    # 获取input_file和output_file的字节数
    input_size=$(stat -c%s "$input_file")
    code=$?
    [[ $code -ne 0 ]] && echo "[ERROR] get_file_size, filename:$input_file" >> ${WORK_DIR}/custom-out.txt && exit $code

    output_size=$(stat -c%s "$output_file")
    code=$?
    [[ $code -ne 0 ]] && echo "[ERROR] get_file_size, filename:$output_file" >> ${WORK_DIR}/custom-out.txt && exit $code

    echo "Input file:$input_file, size:$input_size" >> ${WORK_DIR}/custom-out.txt
    echo "Output file:$output_file, size:$output_size" >> ${WORK_DIR}/custom-out.txt

    # 计算压缩比
    compression_ratio=$(($output_size * 100 / $input_size))
    echo "compression_ratio: $compression_ratio %" >> ${WORK_DIR}/custom-out.txt

    # 判断压缩比是否异常，压缩比在：4..10之间
    if [[ $compression_ratio -ge 10 && $compression_ratio -le 25 ]]; then
        return 0
    else
        echo "[Warn] Possible abnormal compression, filename:$1." >> ${WORK_DIR}/custom-out.txt
        return 1
    fi
    # # Check if compression ratio is less than 4
    # if [ ! $(($input_size / 4)) -le $output_size ]; then
    #     echo "[Warn] Possible abnormal compression. Input file $1 is 4 times larger than output file."  >> ${WORK_DIR}/custom-out.txt
    # fi

    # # Check if compression ratio is less than 8
    # if [ ! $(($input_size / 8)) -le $output_size ]; then
    #     echo "[Error] Compression ratio exceeds normal limits. Exiting with error code 93." >> ${WORK_DIR}/custom-out.txt
    #     return 93
    # fi
}
