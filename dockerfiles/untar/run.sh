#!/bin/bash
#传参：tar文件路径
result=$(echo "$1" | awk -F "%" '{print $2}')
echo $result
input_file="/local"${SOURCE_URL}/$result
InputBytes=$(stat --printf="%s" ${input_file})
echo $InputBytes
#获取文件名
in_file=$(basename ${input_file})
cd /work
#复制文件到当前目录
cp ${input_file} .
#path=$(dirname "$1")
#work_1是暂存解压后的dat文件，在zst压缩过程会删除
#如果文件夹不存在，则创建文件夹
if [ ! -d "work_1" ]; then
mkdir work_1
fi
tar -xf ${in_file} -C /work/work_1
#tar -xvf $1 -C $path/work_1
exit_code=$?
output_dir="/local"$OUTPUT_URL     # 指定输出目录
if [ ! -d "$output_dir" ]; then
  # 目录不存在，使用mkdir命令创建多级目录
  mkdir -p "$output_dir"
else
  echo "多级目录已存在"
fi
rm -f ${in_file}
cd work_1
for myfile in ./*
do
filename=$(basename "$myfile")
zstd --rm "$myfile" -o "${output_dir}/${filename}.zst"
exit_code=$?
groupfilename="zst,${output_dir}/${filename}.zst"
#传给分组
echo "data-grouping-fits,${groupfilename}" >> /work/messages.txt
#python3 /app/bin/run.py $groupfilename
exit_code=$?
done
echo '{
     "inputBytes":'${InputBytes}'
}' > /work/task-exec.json
exit ${exit_code}
