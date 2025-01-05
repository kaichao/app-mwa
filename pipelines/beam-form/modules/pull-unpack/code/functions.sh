#!/bin/bash

# 检查输出文件是否有效？
function post_check() {
    dataset=$1
    ch=$2
    begin=$3
    end=$4
    target_dir=$5

    echo "[post_check]:" >> ${WORK_DIR}/custom-out.txt
    
    # 初始化字节数数组
    sizes=()
    for ((n=$begin; n<=$end; n++)); do
        filename_r="${dataset}_${n}_ch${ch}.dat"
        filename="${target_dir}/$filename_r"
        if [[ -f "$filename" ]]; then
            sizes+=( $(stat -c%s "$filename") )
            echo "file: $filename_r,\t bytes: $(stat -c%s "$filename")" >> ${WORK_DIR}/custom-out.txt
        else
            echo "[ERROR] post_check file $filename_r not exists!" >> ${WORK_DIR}/custom-out.txt
            exit 101
        fi
    done

    # 获取文件数量
    num_files=${#sizes[@]}
    # 计算均值
    mean=0
    for size in "${sizes[@]}"; do
        mean=$((mean + size))
    done
    mean=$((mean / num_files))
    echo " $num_files files, average bytes: $mean" >> ${WORK_DIR}/custom-out.txt

    # 检查除了最后一个文件外其他文件字节数是否一致
    all_equal=true
    for (( i=0; i<num_files-1; i++ )); do
        if [[ "${sizes[i]}" -ne "${sizes[0]}" ]]; then
            all_equal=false
            break
        fi
    done

    if $all_equal; then
        echo "所有文件字节数一致" >> ${WORK_DIR}/custom-out.txt
    else
        echo [ERROR] "文件字节数不一致" >> ${WORK_DIR}/custom-out.txt
        return 102
    fi
}
