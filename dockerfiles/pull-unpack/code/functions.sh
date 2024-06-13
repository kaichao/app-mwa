#!/bin/bash

# 检查输出文件是否有效？
function post_check() {
    begin=$1
    end=$2
    target_dir=$3
    dataset=$4
    ch=$5

    # 初始化字节数数组
    sizes=()
    for ((n=$begin; n<=$end; n++))
    do
        filename="${target_dir}/${dataset}_${n}_ch${ch}.dat"
        if [[ -f "$filename" ]]; then
            sizes+=( $(stat -c%s "$filename") )
        else
            echo "file $filename not exists!"
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

    # 检查除了最后一个文件外其他文件字节数是否一致
    all_equal=true
    for (( i=0; i<num_files-1; i++ )); do
        if [[ "${sizes[i]}" -ne "${sizes[0]}" ]]; then
            all_equal=false
            break
        fi
    done

    if $all_equal; then
        echo "所有文件（除了最后一个）字节数一致"
    else
        echo "文件字节数不一致" >&2

        return 102
    fi

    # 输出均方差
    echo "文件大小的均方差: $std_dev"
}
