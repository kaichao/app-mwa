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
    echo "[ERROR]invalid input message:$1" >&2 && exit 5
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_DAT="/local${LOCAL_OUTPUT_ROOT}/mwa/dat"
else
    DIR_DAT=/data/mwa/dat
fi

if [[ ${arr[0]} == /data* ]]; then
    tar_file="${arr[0]}/${arr[1]}"
else
    tar_file="/local${arr[0]}/${arr[1]}"
fi

tmp_dir="/local/dev/shm/scalebox/copy-unpack"
target_dir="${DIR_DAT}/${dataset}"

echo source_file:$tar_file
echo target_dir:$target_dir

mkdir -p $tmp_dir $target_dir && \
cd $tmp_dir && \
tar xf $tar_file && \
if [ "$KEEP_SOURCE_FILE" = "no" ]; then rm -f tar_file;fi
code=$?
[[ $code -ne 0 ]] && echo "error untar file:$tar_file" >&2 && exit $code

echo "0000" >&2
echo "0000"
echo tmp_dir:$tmp_dir 
ls -l $tmp_dir

echo "0010"
echo "0010" >&2

for f in $(ls *.zst); do
echo "1000,f=$f" >&2
    zstd -d -f --output-dir-flat=$target_dir --rm $f
    code=$?
echo "1001,code=$code" >&2
    if [[ $code -ne 0 ]]; then 
        zstd -d -f --output-dir-flat=$target_dir --rm $f
        code=?
echo "1002,code=$code" >&2
    fi
    [[ $code -ne 0 ]] && echo "error unzstd file:$f" >&2 && exit $code
echo "1003" >&2
done
echo "2222" >&2


# Read error (39) : premature end 

# 删除临时文件
rm -f $tmp_dir/*
cd $target_dir && chmod 644 *.dat

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
