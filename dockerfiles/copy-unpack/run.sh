#!/bin/bash

# /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch121.dat.zst.tar
m=$1

arr=($(echo $m | tr "~" " ")) 

if [[ ${arr[1]} =~ ^([0-9]+)/([0-9]+)_([0-9]+)_ch([0-9]{3}).dat.zst.tar$ ]]; then
    dataset=${BASH_REMATCH[1]}
    begin=${BASH_REMATCH[2]}
    end=${BASH_REMATCH[3]}
    ch=${BASH_REMATCH[4]}
else
    echo "invalid input message:$1" >&2 && exit 5
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_DAT="/local${LOCAL_OUTPUT_ROOT}/mwa/dat"
else
    DIR_DAT=/data/mwa/dat
fi

file_name="/local${arr[0]}/${arr[1]}"
tmp_dir="/local/dev/shm/copy-unpack"
target_dir="${DIR_DAT}/${dataset}"

echo source_file:$file_name
echo target_dir:$target_dir

mkdir -p $tmp_dir $target_dir && \
cd $tmp_dir && \
tar xf $file_name && \
if [ "$KEEP_SOURCE_FILE" = "no" ]; then rm -f file_name;fi && \
zstd -d -f --output-dir-flat=$target_dir --rm *.zst
code=$?

# 删除临时文件
rm -f /local/dev/shm/copy-unpack/*

[[ $code -ne 0 ]] && echo "error copy-unpack file:$f" >&2 && exit $code

# /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch132.dat.zst.tar
# 1257010784/1257010784_1257010801_ch132.dat

# for n in {$begin..$end}; do
#     echo "${dataset}/${dataset}_${n}_ch${ch}.dat" > ${WORK_DIR}/messages.txt
# done

for ((n=$begin; n<=$end; n++))
do
    echo "${dataset}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/messages.txt
done

exit $code
