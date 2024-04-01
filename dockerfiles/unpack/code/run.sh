#!/bin/bash
# 第3部分为batch-index，仅用于避免重复key_message
# 1257010784/1257010786_1257010815_ch109.dat.tar.zst~1257010786_1257010845~b01
arr=($(echo $1 | tr "~" " ")) 

if [[ ${arr[0]} =~ ^([0-9]+)/([0-9]+)_([0-9]+)_ch([0-9]+)\.dat\.tar\.zst$ ]]; then
    dataset="${BASH_REMATCH[1]}"
    begin="${BASH_REMATCH[2]}"
    end="${BASH_REMATCH[3]}"
    ch="${BASH_REMATCH[4]}"
else
    echo "[ERROR] Input does not match :$1" >&2 && exit 5
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_DAT="/local${LOCAL_OUTPUT_ROOT}/mwa/dat"
else
    DIR_DAT=/data/mwa/dat
fi

source_file="/local${LOCAL_INPUT_ROOT}/mwa/tar/${arr[0]}"
target_dir="${DIR_DAT}/${dataset}/ch${ch}/${arr[1]}"

echo source_file:$source_file
echo target_dir:$target_dir

mkdir -p "$target_dir" && cd "$target_dir" && zstd -dc "$source_file" | tar xf -

code=$?
[[ $code -ne 0 ]] && echo "error unpack file:$source_file" >&2 && exit $code

cd $target_dir && chmod 644 *.dat

# /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch132.dat.tar.zst
# 1257010784/1257010784_1257010801_ch132.dat
for ((n=$begin; n<=$end; n++))
do
    echo "${dataset}_${n}_ch${ch}.dat~${arr[2]}" >> ${WORK_DIR}/messages.txt
    # echo "${DIR_DAT}/${dataset}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/output-files.txt
    echo "${target_dir}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/output-files.txt
    echo "output-file: ${target_dir}/${dataset}_${n}_ch${ch}.dat"
done

[ "$KEEP_SOURCE_FILE" = "no" ] && echo "$source_file be removed" && echo $source_file >> $WORK_DIR/removed-files.txt
echo $source_file >> ${WORK_DIR}/input-files.txt

exit $code
