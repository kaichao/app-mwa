#!/bin/bash

# 指定要查找的目录路径
# RESULT_DIR=astro@10.100.1.30:10022/data2/mydata/mwa/png
# LOCAL_RESULT_DIR="/data2/mydata/mwa/png/1301240224-240926"

# LOCAL_RESULT_DIR=/cluster_data_root/mwa/png
# echo $LOCAL_RESULT_DIR > $WORK_DIR/custom-out.txt
tmp_file="tmppoint.txt"

dataset=$1
output_file=$2
PB=$3
PE=$4

# 清空输出文件
> ${WORK_DIR}/${tmp_file}
> ${WORK_DIR}/${output_file}

# 如果设置了环境变量$FIX_MISSING=yes，则执行以下内容
if [[ "$FIX_MISSING" == "yes" ]]; then

    # 获取实际存在的子目录名并存储在数组中
    echo "/local/${LOCAL_RESULT_DIR}/${dataset}"
    ls -d /local/${LOCAL_RESULT_DIR}/${dataset}/p* > $WORK_DIR/custom-out.txt
    existing_dirs=($(ls -d /local/${LOCAL_RESULT_DIR}/${dataset}/p* 2>/dev/null | xargs -n 1 basename))

    # echo $existing_dirs
    # 循环检查 PB 至 PE 之间的目录是否存在
    for i in $(seq $PB $PE); do
        dir_name=$(printf "p%05d" "$i")
        if [[ ! " ${existing_dirs[@]} " =~ " ${dir_name} " ]]; then
            echo "$dir_name" >> "${WORK_DIR}/${tmp_file}"
        fi
    done

    # 遍历目录下的所有子目录
    for dir in ${existing_dirs[@]}; do
        # 统计子目录中的文件数量（包括隐藏文件）
        file_count=$(find "/local/${LOCAL_RESULT_DIR}/${dataset}/${dir}" -type f | wc -l)
        small_files=$(find "/local/${LOCAL_RESULT_DIR}/${dataset}/${dir}" -type f -size -100000c)
        # 如果文件数量小于7，输出子目录名称到文件
        if [[ "$file_count" -lt $MAX_LINENUM ]]; then
            echo "$(basename "$dir")" >> "${WORK_DIR}/${tmp_file}"
        # 如果找到小文件，则输出子目录名称到文件
        elif [[ -n "$small_files" ]]; then
            echo "$(basename "$dir")" >> "${WORK_DIR}/${tmp_file}"
        fi
    done

    cat ${WORK_DIR}/${tmp_file} | sort > ${WORK_DIR}/${output_file}
    rm ${WORK_DIR}/${tmp_file}
else
    # 将PB到PE之间的目录名称输出到文件
    for i in $(seq $PB $PE); do
        dir_name=$(printf "p%05d" "$i")
        echo "$dir_name" >> "${WORK_DIR}/${output_file}"
    done
fi
echo "pointings filtered into ${WORK_DIR}/${output_file}"

