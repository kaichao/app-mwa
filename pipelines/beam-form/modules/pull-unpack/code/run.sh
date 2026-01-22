#!/bin/bash

source functions.sh
source /app/share/bin/functions.sh
source $(dirname $0)/functions.sh

# 1266932744/p00001_00960/1266932986_1266933025_ch118.dat.tar.zst
# s=$1
# 删除第一个'/'之前的部分，得到 1257010784/1257010786_1257010815_ch109.dat.tar.zst
# m="${s#*/}"
# pointing="${s%%/*}"

if [[ $1 =~ ^([0-9]+)/((p[0-9]+_[0-9]+)/)?(([0-9]+)_([0-9]+)_ch([0-9]+)\.dat\.tar\.zst)$ ]]; then
    dataset="${BASH_REMATCH[1]}"
    pointing_path="${BASH_REMATCH[3]}"
    file_name="${BASH_REMATCH[4]}"
    begin="${BASH_REMATCH[5]}"
    end="${BASH_REMATCH[6]}"
    ch="${BASH_REMATCH[7]}"
else
    echo "[ERROR] Input does not match :$1" >&2 && exit 5
fi

echo "pointing_path=$pointing_path"
# jump_servers=$(get_header "$2" "jump_servers")
# jump_servers_option=""
# if [ $jump_servers ]; then
#     jump_servers_option="-J '${jump_servers}' "
# fi
# ssh_args="-T -c aes128-gcm@openssh.com -o Compression=no -x ${jump_servers_option}"

# target_url is local-dir
target_url=$(get_header "$2" "target_url")
target_subdir=$(get_header "$2" "target_subdir")
if [[ "$target_url" == /* ]]; then
    target_dir="${target_url}/${target_subdir}"
else
    target_dir="${LOCAL_TMPDIR}/${target_url}/${target_subdir}"
fi
#target_dir=$(get_host_path "${target_url}/${target_subdir}")

source_url=$(get_header "$2" "source_url")
source_mode=$(get_mode "$source_url")
source_dir=$(get_data_root "$source_url")

bw_limit=$(get_header "$2" "bw_limit")
# BW_LIMIT  "500k"/"1m"
# 设置--touch，将文件更新时间更新为当前时间，以免在/tmp中被删除
if [ -n "$bw_limit" ]; then
    # 已设置
    cmd_part="pv -q -L ${bw_limit}|zstd -d | tar --touch -xvf -"
else
    cmd_part="zstd -d | tar --touch -xvf -"
fi
# pv -L 500k source_file > destination_file
if [ "$source_mode" = "LOCAL" ]; then
    source_dir=$(get_host_path $source_dir)
    source_file="${source_dir}/mwa/tar/$dataset/$file_name"
    cmd="cat $source_file | ${cmd_part}"
else
    # SSH
    ssh_cmd=$(get_ssh_cmd "$2" "source_url" "source_jump")
    cmd="$ssh_cmd \"cat ${source_dir}/mwa/tar/$dataset/$file_name\" - | ${cmd_part}"
fi

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

echo "source_url:$source_url, source_dir:$source_dir" >> ${WORK_DIR}/auxout.txt
echo "target_url:$target_url, target_dir:$target_dir, target_subdir:$target_subdir" >> ${WORK_DIR}/auxout.txt
echo "cmd:$cmd" >> ${WORK_DIR}/auxout.txt

echo "message:$1" >> ${WORK_DIR}/auxout.txt

# cmd="ssh -p ${ssh_port} ${ssh_args} ${ssh_host} \"cat ${source_dir}/$m\" - | zstd -d | tar -xvf -"
mkdir -p ${target_dir} && cd ${target_dir} && eval $cmd 
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]exit after pull-unpack, error_code:$code"  >> ${WORK_DIR}/auxout.txt && exit $code

chmod 644 *.dat

# 检查输出文件是否完整
post_check $dataset $ch $begin $end $target_dir
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]exit after post-check output files, exit_code:$code"  >> ${WORK_DIR}/auxout.txt && exit $code

for ((n=$begin; n<=$end; n++))
do
    echo "${target_dir}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/output-files.txt
done

echo "$1" >> ${WORK_DIR}/messages.txt

if [ "$source_mode" = "LOCAL" ]; then
    echo "${source_file}" >> ${WORK_DIR}/input-files.txt
    keep_source_file=$(get_header "$2" "keep_source_file")
    if [ "$keep_source_file" = "no" ]; then
        echo "${source_file}" > ${WORK_DIR}/removed-files.txt
    fi
fi

exit $code
