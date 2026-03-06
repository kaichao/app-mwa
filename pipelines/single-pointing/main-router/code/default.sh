#!/bin/bash

current_dir=`dirname $0`

filename="$current_dir/my-pointings.txt"
# 读取文件的第一行
line=$(head -n 1 "$filename")

$current_dir/process-single-pointing.sh $line

# 逐行读取文件内容
# while IFS= read -r line
# do
#   echo "$line"
#   $current_dir/process-single-pointing.sh $line
# done < "$filename"

exit 0
