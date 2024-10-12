#!/bin/bash
#传参：tar文件路径
#result=$(echo "$1" | awk -F "~" '{print $2}')
result="${1#*~}"
#tar1266932744/1266932744_1266935421x.tar.zst

echo "result:"$result
input_file="/local"${SOURCE_URL}/$result
InputBytes=$(stat --printf="%s" ${input_file})
exit_code=$?
echo $InputBytes
#获取文件名
in_file=$(basename ${input_file})
cd /work
#复制文件到当前目录
cp ${input_file} .
exit_code=$?
#path=$(dirname "$1")
#work_1是暂存解压后的dat文件，在zst压缩过程会删除
#如果文件夹不存在，则创建文件夹
if [ ! -d "work_1" ]; then
mkdir work_1
fi
zstd -d -c ${in_file} | tar -xvf - --transform 's|.*/||' -C /work/work_1
#tar -xvf ${in_file} --transform 's|.*/||' -C /work/work_1

#tar -xf ${in_file} -C /work/work_1
#tar -xvf $1 -C $path/work_1
exit_code=$?
if [ $exit_code -ne 0 ]; then
    exit ${exit_code}
fi
output_dir="/local/data/mwa"     # 指定输出目录
dat_output_dir="/local"${OUTPUT_URL}
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

#mv ${filename} ${output_dir}
#send-message $filename
if [[ $filename == *"ch"* ]]; then
    mv ${filename} ${output_dir}
    send-message ${filename}
else
    mv ${filename} ${dat_output_dir}
fi

#zstd --rm "$myfile" -o "${output_dir}/${filename}.zst"
#exit_code=$?
done
#send-message $1
echo '{
     "inputBytes":'${InputBytes}'
}' > /work/task-exec.json
exit ${exit_code}
