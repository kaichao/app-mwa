#!/bin/bash

# /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch121.dat.zst.tar~1257010786_1257010875
# 1257010784/1257010786_1257010815_ch109.dat.zst.tar
m=$1

arr=($(echo $m | tr "~" " ")) 

if [[ ${arr[0]} =~ ^([0-9]+)/([0-9]+)_([0-9]+)_ch([0-9]{3}).dat.zst.tar$ ]]; then
    dataset=${BASH_REMATCH[1]}
    begin=${BASH_REMATCH[2]}
    end=${BASH_REMATCH[3]}
    ch=${BASH_REMATCH[4]}
else
    echo "[ERROR] Invalid input message:$1" >&2 && exit 5
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_DAT="/local${LOCAL_OUTPUT_ROOT}/mwa/dat"
else
    DIR_DAT=/data/mwa/dat
fi


tar_file="/local${LOCAL_INPUT_ROOT}/mwa/tar/${arr[0]}"
# if [[ ${arr[0]} == /data* ]]; then
#     tar_file="${arr[0]}/${arr[1]}"
# else
#     tar_file="/local${arr[0]}/${arr[1]}"
# fi

tmp_dir="/local/dev/shm/scalebox/copy-unpack"
target_dir="${DIR_DAT}/${dataset}/ch${ch}/${arr[1]}"

echo source_file:$tar_file
echo target_dir:$target_dir

[ "$KEEP_SOURCE_FILE" = "no" ] && echo $tar_file >> $WORK_DIR/removed-files.txt

mkdir -p $tmp_dir $target_dir && cd $tmp_dir && tar xf $tar_file
code=$?
[[ $code -ne 0 ]] && echo "error untar file:$tar_file" >&2 && exit $code

echo "[INFO]tmp_dir=$tmp_dir"
ls -l $tmp_dir/*.zst

for f in $(ls *.zst); do
    zstd -d -f --output-dir-flat=$target_dir --rm $f
    code=$?
    if [[ $code -ne 0 ]]; then 
        zstd -d -f --output-dir-flat=$target_dir --rm $f
        code=?
    fi
    [[ $code -ne 0 ]] && echo "error unzstd file:$f" >&2 && exit $code
done

# Read error (39) : premature end 

# 删除临时文件
rm -f $tmp_dir/*
cd $target_dir && chmod 644 *.dat

[[ $code -ne 0 ]] && echo "error copy-unpack file:$f" >&2 && exit $code

# /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch132.dat.zst.tar
# 1257010784/1257010784_1257010801_ch132.dat
for ((n=$begin; n<=$end; n++))
do
    echo "${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/messages.txt
    # echo "${DIR_DAT}/${dataset}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/output-files.txt
    echo "${target_dir}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/output-files.txt
    echo "output-file: ${target_dir}/${dataset}_${n}_ch${ch}.dat"
done

echo $tar_file >> ${WORK_DIR}/input-files.txt

exit $code
