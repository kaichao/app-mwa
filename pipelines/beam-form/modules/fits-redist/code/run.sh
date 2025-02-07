#!/usr/bin/env bash

source functions.sh
# source $(dirname $0)/functions.sh

if [ $LOCAL_INPUT_ROOT ]; then
    DIR_1CHX=$(get_host_path "${LOCAL_INPUT_ROOT}/mwa/1chx")
else
    DIR_1CHX=/cluster_data_root/mwa/1chx
fi

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_1CHZ=$(get_host_path "${LOCAL_OUTPUT_ROOT}/mwa/1chz")
else
    DIR_1CHZ=/cluster_data_root/mwa/1chz
fi

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

cd ${DIR_1CHX}/$1
echo pwd:$PWD >> ${WORK_DIR}/custom-out.txt
# 使用 mapfile 将当前目录下所有文件名读取到数组中
mapfile -t arr_files < <(ls -p | grep -v /)

echo files:${#arr_files[@]},hosts:${#arr_hosts[@]} >> ${WORK_DIR}/custom-out.txt

# 检查 arr_files 和 arr_hosts 的长度是否相同
if [ ${#arr_files[@]} -ne ${#arr_hosts[@]} ]; then
    echo "【ERROR] arr_files and arr_hosts have different lengths!" >&2
    exit 1
fi

export SOURCE_URL=${DIR_1CHZ}

target_user=${TARGET_USER:-root}
target_port=${TARGET_PORT:-22}
target_root=${TARGET_ROOT:-${CLUSTER_DATA_ROOT}/mwa/1chz}

for i in "${!arr_files[@]}"; do
    echo "file: ${arr_files[i]}, host: ${arr_hosts[i]}"
    f=${arr_files[i]}
    p="${f%%.fits.zst}"
    m="${ds}/${p}/${t_label}/${ch}.fits.zst"
    file_1chz="${DIR_1CHZ}/$m"
    mkdir -p "$(dirname $file_1chz)"
    if [ "$KEEP_SOURCE_FILE" == "yes" ]; then
        cp -f $f $file_1chz
    else
        mv -f $f $file_1chz
    fi

    if [ ${arr_hosts[i]} == "localhost" ]; then
        continue
    fi
    # export TARGET_URL=root@${arr_hosts[i]}:22${DIR_1CHZ}
    export TARGET_URL=${target_user}@${arr_hosts[i]}:${target_port}${target_root}
    # 循环调用/app/share/bin/run.sh，分发文件
    eval "/app/share/bin/run.sh '$m' '$2'"
    code=$?
    [[ $code -ne 0 ]] && echo "[ERROR]exit after file-copy, error_code:$code"  >> ${WORK_DIR}/custom-out.txt && exit $code
done

# 删除 /app/share/bin/run.sh 调用产生的消息
rm -f ${WORK_DIR}/messages.txt
echo "$1" > ${WORK_DIR}/messages.txt

exit 0
