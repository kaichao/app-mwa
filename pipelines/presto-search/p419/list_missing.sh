#!/bin/bash

# 指定要查找的目录路径
base_dir="/data2/mydata/mwa/png/1301240224-240926"
tmp_file="tmppoint.txt"
output_file="pointings.txt"

# 清空输出文件
> "$tmp_file"

# 获取实际存在的子目录名并存储在数组中
existing_dirs=($(ls -d "$base_dir"/p0* 2>/dev/null | xargs -n 1 basename))

# 循环检查 p00960 至 p01920 之间的目录是否存在
for i in $(seq 3841 4800); do
    dir_name=$(printf "p%05d" "$i")
    if [[ ! " ${existing_dirs[@]} " =~ " ${dir_name} " ]]; then
        echo "$dir_name" >> "$tmp_file"
    fi
done

# 遍历目录下的所有子目录
for dir in "$base_dir"/*/; do
    # 统计子目录中的文件数量（包括隐藏文件）
    file_count=$(find "$dir" -type f | wc -l)
    small_files=$(find "$dir" -type f -size -100000c)
    # 如果文件数量小于7，输出子目录名称到文件
    if [[ "$file_count" -lt 7 ]]; then
        echo "$(basename "$dir")" >> "$tmp_file"
    # 如果找到小文件，则输出子目录名称到文件
    elif [[ -n "$small_files" ]]; then
        echo "$(basename "$dir")" >> "$tmp_file"
    fi
done

cat $tmp_file | sort > $output_file
rm $tmp_file

echo "缺失目录已写入 $output_file"

