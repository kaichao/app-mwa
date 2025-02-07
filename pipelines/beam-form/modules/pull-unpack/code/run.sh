#!/bin/bash

source functions.sh
source /app/share/bin/functions.sh
source $(dirname $0)/functions.sh

# 1257010784/1257010786_1257010815_ch109.dat.tar.zst
m=$1

# remove last characters ~b01
#m="${m0%~*}"
#batch="${m0##*~}"

if [[ $m =~ ^([0-9]+)/([0-9]+)_([0-9]+)_ch([0-9]+)\.dat\.tar\.zst$ ]]; then
    dataset="${BASH_REMATCH[1]}"
    begin="${BASH_REMATCH[2]}"
    end="${BASH_REMATCH[3]}"
    ch="${BASH_REMATCH[4]}"
else
    echo "[ERROR] Input does not match :$1" >&2 && exit 5
fi

# jump_servers=$(get_header "$2" "jump_servers")
# jump_servers_option=""
# if [ $jump_servers ]; then
#     jump_servers_option="-J '${jump_servers}' "
# fi
# ssh_args="-T -c aes128-gcm@openssh.com -o Compression=no -x ${jump_servers_option}"

# target_url is local-dir
target_url=$(get_header "$2" "target_url")
target_subdir=$(get_header "$2" "target_subdir")
target_dir=$(get_host_path "${target_url}/${target_subdir}")

source_url=$(get_header "$2" "source_url")
source_mode=$(get_mode "$source_url")
source_dir=$(get_data_root "$source_url")

# BW_LIMIT  "500k"/"1m"
# 设置--touch，将文件更新时间更新为当前时间，以免在/tmp中被删除
if [ -n "$BW_LIMIT" ]; then
    # 已设置
    cmd_part="pv -L ${BW_LIMIT}|zstd -d | tar --touch -xvf -"
else
    cmd_part="zstd -d | tar --touch -xvf -"
fi
# pv -L 500k source_file > destination_file
if [ "$source_mode" = "LOCAL" ]; then
    source_dir=$(get_host_path $source_dir)
    cmd="cat ${source_dir}/$m | ${cmd_part}"
else
    # RSYNC-OVER-SSH
    ssh_cmd=$(get_ssh_cmd "$2" "source_url" "source_jump_servers")
    cmd="$ssh_cmd \"cat ${source_dir}/$m\" - | ${cmd_part}"
fi

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

echo "source_url:$source_url" >> ${WORK_DIR}/custom-out.txt
echo "source_dir:$source_dir" >> ${WORK_DIR}/custom-out.txt
echo "target_url:$target_url" >> ${WORK_DIR}/custom-out.txt
echo "target_subdir:$target_subdir" >> ${WORK_DIR}/custom-out.txt
echo "target_dir:$target_dir" >> ${WORK_DIR}/custom-out.txt
echo "cmd:$cmd" >> ${WORK_DIR}/custom-out.txt

echo "message:$m" >> ${WORK_DIR}/custom-out.txt

# cmd="ssh -p ${ssh_port} ${ssh_args} ${ssh_host} \"cat ${source_dir}/$m\" - | zstd -d | tar -xvf -"
mkdir -p ${target_dir} && cd ${target_dir} && eval $cmd 
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]exit after pull-unpack, error_code:$code"  >> ${WORK_DIR}/custom-out.txt && exit $code

chmod 644 *.dat

# 检查输出文件是否完整
post_check $dataset $ch $begin $end $target_dir
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]exit after post-check output files, exit_code:$code"  >> ${WORK_DIR}/custom-out.txt && exit $code

# 消息加上批次号，以免在多批次处理过程中，在message-router中有同名冲突
for ((n=$begin; n<=$end; n++))
do
    echo "${target_dir}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/output-files.txt
done

echo "$1" >> ${WORK_DIR}/messages.txt

if [ "$source_mode" = "LOCAL" ]; then
    echo "${source_dir}/$m" > ${WORK_DIR}/output-files.txt
    if [ "$KEEP_SOURCE_FILE" != "yes" ]; then
        echo "${source_dir}/$m" > ${WORK_DIR}/removed-files.txt
    fi
fi

exit $code
