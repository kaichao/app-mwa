#!/usr/bin/env bash

source functions.sh
# source $(dirname $0)/functions.sh

if [ $INPUT_ROOT ]; then
    dir_1chx=$(get_host_path "${INPUT_ROOT}/mydata/mwa/1chx")
else
    dir_1chx=/cluster_data_root/mwa/1chx
fi

if [ $OUTPUT_ROOT ]; then
    dir_1chy=$(get_host_path "${OUTPUT_ROOT}/mydata/mwa/1chy")
else
    dir_1chy=/cluster_data_root/mwa/1chy
fi

# if [ $OUTPUT_ROOT ]; then
#     DIR_1CHZ=$(get_host_path "${OUTPUT_ROOT}/mwa/1chz")
# else
#     DIR_1CHZ=/cluster_data_root/mwa/1chz
# fi

# 输入消息： 1257617424/p00001_00024/t1257617426_1257617505/ch113
regex='^([0-9]+)/(p[0-9]+_[0-9]+)/(t[0-9]+_[0-9]+)/(ch[0-9]+)$'
if [[ $1 =~ $regex ]]; then
    # echo "${BASH_REMATCH[1]}"  # 1257617424
    # echo "${BASH_REMATCH[2]}"  # p00001_00024
    # echo "${BASH_REMATCH[3]}"  # t1257617426_1257617505
    # echo "${BASH_REMATCH[4]}"  # ch109
    ds="${BASH_REMATCH[1]}"
    t_label="${BASH_REMATCH[3]}"
    ch="${BASH_REMATCH[4]}"
else
    echo "[ERROR]Invalid Format, filename:$1" >&2
    exit 81
fi

target_hosts=$(get_header "$2" "target_hosts")
# 设置 IFS 为逗号
IFS=',' read -r -a arr_hosts <<< "$target_hosts"

cd "${dir_1chx}/$1"
echo pwd:$PWD >> ${WORK_DIR}/custom-out.txt
# 使用 mapfile 将当前目录下所有文件名读取到数组中
mapfile -t arr_files < <(ls -p | grep -v /)

echo files:${#arr_files[@]},hosts:${#arr_hosts[@]} >> ${WORK_DIR}/custom-out.txt

# 检查 arr_files 和 arr_hosts 的长度是否相同
if [ ${#arr_files[@]} -ne ${#arr_hosts[@]} ]; then
    echo "【ERROR] arr_files and arr_hosts have different lengths!" >&2
    exit 1
fi

export SOURCE_URL=${LOCAL_SHMDIR}

target_user=${TARGET_USER:-root}
target_port=${TARGET_PORT:-22}
# target_root=${TARGET_ROOT:-${CLUSTER_DATA_ROOT}}
# target_dir="${target_root}/mwa/1chz"
# echo "target_dir:$target_dir"

ret_code=0
for i in "${!arr_files[@]}"; do
    echo "file: ${arr_files[i]}, host: ${arr_hosts[i]}"
    f=${arr_files[i]}
    p="${f%%.fits.zst}"
    fn="${ds}/${p}/${t_label}/${ch}.fits.zst"
    file_1chy="${dir_1chy}/$fn"
    mkdir -p "$(dirname $file_1chy)"
    if [ "$KEEP_SOURCE_FILE" == "yes" ]; then
        cp -f $f $file_1chy
    else
        mv -f $f $file_1chy
    fi
    echo "source=$f,target=$file_1chy,m=$m,host:${arr_hosts[i]}" >> ${WORK_DIR}/custom-out.txt

    if [ ${arr_hosts[i]} == "localhost" ]; then
        continue
    fi
#    export TARGET_URL=${target_user}@${arr_hosts[i]}:${target_port}${target_dir}
    export TARGET_URL=${target_user}@${arr_hosts[i]}:${target_port}${LOCAL_SHMDIR}
    # 循环调用/app/share/bin/run.sh，分发文件
    m="mydata/mwa/1chy/$fn"
    eval "/app/share/bin/run.sh '$m' '$2'"
    code=$?
    if [[ $code -ne 0 ]]; then
        ret_code=$code
    fi
    echo "after file-copy,file=$m, ret_code=$ret_code, code:$code" >> "${WORK_DIR}/custom-out.txt"
done

echo removed-files: >> ${WORK_DIR}/custom-out.txt
cat ${WORK_DIR}/removed-files.txt >> ${WORK_DIR}/custom-out.txt
cat "ret_code:$ret_code" >> ${WORK_DIR}/custom-out.txt

# 删除 /app/share/bin/run.sh 调用产生的消息
rm -f ${WORK_DIR}/messages.txt
if [[ $ret_code -ne 0 ]]; then
    rm -f ${WORK_DIR}/removed-files.txt
    # TODO : 1chy -> 1chx

    exit $ret_code
fi

echo "$1" > ${WORK_DIR}/messages.txt

exit 0
