#!/bin/bash
echo $1
#path为文件的存储路径，st为数据集时间戳，t1为打包数据的起始时间戳，t2是终止时间戳，c是通道

pathtar="/local"$OUT_URL

if [ ! -d "$pathtar" ]; then
  # 目录不存在，使用mkdir命令创建多级目录
  mkdir -p "$pathtar"
else
  echo "多级目录已存在"
fi
#path1=$(dirname "$path")
#暂存的路径用容器自身的，不能用共享存储，多个slot一起运行时使用共享存储路径内文件会出问题
path1='/work'
echo "path1:"$path1

if [ ! -d "$path1/work_final" ]; then
mkdir $path1/work_final
echo "文件夹不存在"
exit_code=$?
fi
#如果非空，删除$path1/work_final下文件，保证为空文件夹
if [ "$(ls -A $path1/work_final)" ];
then 
rm "$path1/work_final"/*
fi

input_string=$1
IFS=','
first_str=$(echo "$input_string" | cut -d ',' -f1)
echo "First String: $first_str"
# 从第一个逗号前截取数字部分
first_number=$(echo "$first_str" | awk -F '_' '{print $2}')
echo "First Number: $first_number"
# 读取最后一个子字符串
last_str=$(echo "${input_string##*,}")
echo "Last String: $last_str"

# 从最后一个逗号后截取数字部分和最后一个逗号后的字符串部分
last_number=$(echo "$last_str" | awk -F '_' '{print $2}')
suffix=$(echo "$last_str" | awk -F '_' '{print $3}' | cut -d '.' -f1)
echo "Last Number: $last_number"
echo "Suffix: $suffix"

for str in $input_string; do
    cp "$str" "$path1/work_final"
    exit_code=$?
done

cd $path1
#打包后的文件放回原目录下，如果需要放到其他的文件夹，改动$path/${t1}_${t2}_ch${c}.dat.zst.tar前面路径部分

tar  -cvf "$pathtar/${first_number}_${last_number}_${suffix}.dat.zst.tar" -C $path1/work_final .

#发给对象存储
send-message ${first_number}_${last_number}_${suffix}.dat.zst.tar
exit_code=$?
#清空work_final文件夹以便后续使用
rm "$path1/work_final"/*
exit ${exit_code}