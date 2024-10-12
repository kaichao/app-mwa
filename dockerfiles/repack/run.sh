#!/bin/bash
echo $1
#$1为'{datasetid}/{start}_{end}/ch{i}'格式的内容
#测试$1=1253471952/1253471954_1253471955/ch110
#path为文件的存储路径，st为数据集时间戳，t1为打包数据的起始时间戳，t2是终止时间戳，c是通道
message=$1
#input为dat数据文件存储路径

#SOURCE_URL='/home/zhouhan/zhou/test/data'
#暂存的路径用容器自身的，不能用共享存储，多个slot一起运行时使用共享存储路径内文件会出问题
path1="/local"${OUT_URL}
#dat_dir为待打包文件的暂存文件夹
dat_dir="/work/dat_dir"

dat_a="${path1}/dat_dir"
if [ ! -d "$dat_dir" ]; then
mkdir $dat_dir
echo "文件夹不存在"
exit_code=$?
fi
#如果非空，删除$dat_dir下文件，保证为空文件夹
if [ "$(ls -A $dat_dir)" ];
then 
rm "$dat_dir"/*
fi
echo "dat_dir:"$dat_dir
IFS='/'
# 从第一个/前截取数字部分为数据集ID
datasetid=$(echo "$message" | awk -F '/' '{print $1}')
echo "DatasetId: $datasetid"
# 从第一个/后第二个/截取数字部分为起始时间戳和终止时间戳
second_str=$(echo "$message" | awk -F '/' '{print $2}')
star_st=$(echo "$second_str" | awk -F '_' '{print $1}')
end_st=$(echo "$second_str" | awk -F '_' '{print $2}')
echo "star: ${star_st}"
echo "end: ${end_st}"
# 读取最后一个子字符串，获取通道号
num_of_ch=$(echo "${message##*/}" | awk -F 'ch' '{print $2}')
echo "number of channel: ${num_of_ch}"
# 进行待打包文件的迁移
for ((i=${star_st}; i<=${end_st}; i++))
do
    datname=${datasetid}_${i}_ch${num_of_ch}.dat
    mv "/local${SOURCE_URL}/${datname}" "$dat_dir"
    #cp "${SOURCE_URL}/${datname}" "$dat_dir"
    exit_code=$?
done

cd $path1
echo "dat_dir:""$dat_dir"

file_count=$(find "$dat_dir" -type f | wc -l)
echo "file_count:"$file_count
allnum=$((end_st - star_st+1))
if [ "$file_count" -eq "$allnum" ]; then
    tarfile=${path1}/${star_st}_${end_st}_ch${num_of_ch}.dat.tar
    #打包后的文件放回/work目录下
    tar -cvf - -C "$dat_dir" . | zstd --rm -o "${tarfile}.zst"
    #tar  -cvf "$tarfile" -C "$dat_dir" .
    # echo "打包"$dat_dir"文件为"${tarfile}
    #zstd --rm "$tarfile" -o "${tarfile}.zst"
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        rm "$dat_dir"/*
        tarfilename=${star_st}_${end_st}_ch${num_of_ch}.dat.tar.zst
        send-message ${tarfilename}
    fi
    #rm "$dat_dir"/*
    exit ${exit_code}
else
    num=$allnum-$file_count
    echo "两个数字不相等:"$num
    exit 1
fi

