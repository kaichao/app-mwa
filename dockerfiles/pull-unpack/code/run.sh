#!/bin/bash

source functions.sh

# 1257010784/1257010786_1257010815_ch109.dat.tar.zst~b01
m0=$1

# remove last characters ~b01
m="${m0%~*}"
batch="${m0##*~}"

if [[ $m =~ ^([0-9]+)/([0-9]+)_([0-9]+)_ch([0-9]+)\.dat\.tar\.zst$ ]]; then
    dataset="${BASH_REMATCH[1]}"
    begin="${BASH_REMATCH[2]}"
    end="${BASH_REMATCH[3]}"
    ch="${BASH_REMATCH[4]}"
else
    echo "[ERROR] Input does not match :$1" >&2 && exit 5
fi

jump_servers=$(get_parameter "$2" "jump_servers")
jump_servers_option=""
if [ $jump_servers ]; then
    jump_servers_option="-J '${jump_servers}' "
fi
ssh_args="-T -c aes128-gcm@openssh.com -o Compression=no -x ${jump_servers_option}"
#ssh_args="-c aes128-gcm@openssh.com -o Compression=no -x ${jump_servers_option}"

echo "jump_servers:$jump_servers"

source_url=$(get_parameter "$2" "source_url")
target_url=$(get_parameter "$2" "target_url")

# user@10.1.1.1:10022:/raid0/1301240224
# 冒号数量
colon_count=$(echo "$source_url" | awk -F':' '{print NF-1}')
echo "colon_count:$colon_count" >> ${WORK_DIR}/custom-out.txt
if [ "$colon_count" -ge 2 ]; then
    IFS=':' read -r ssh_host ssh_port source_dir <<< ${source_url}
else
    IFS=':' read -r ssh_host source_dir <<< ${source_url}
    ssh_port=22
fi


date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

echo "source_url:$source_url" >> ${WORK_DIR}/custom-out.txt
echo "source_dir:$source_dir" >> ${WORK_DIR}/custom-out.txt
echo "target_url:$target_url" >> ${WORK_DIR}/custom-out.txt
echo "ssh_host:$ssh_host" >> ${WORK_DIR}/custom-out.txt
echo "ssh_port:$ssh_port" >> ${WORK_DIR}/custom-out.txt
echo "ssh_args:$ssh_args" >> ${WORK_DIR}/custom-out.txt
echo "message:$m" >> ${WORK_DIR}/custom-out.txt
target_dir="/local${target_url}"

cmd="ssh -p ${ssh_port} ${ssh_args} ${ssh_host} \"cat ${source_dir}/$m\" - | zstd -d | tar -xvf -"
mkdir -p ${target_dir} \
    && cd ${target_dir}
eval $cmd
code=$?

[[ $code -ne 0 ]] && echo "exit after pull-unpack, error_code:$code" >&2 && exit $code

# 消息加上批次号，以免在多批次处理过程中，在message-router中有同名冲突
for ((n=$begin; n<=$end; n++))
do
    echo "${dataset}_${n}_ch${ch}.dat~${batch}" >> ${WORK_DIR}/messages.txt
    # echo "${DIR_DAT}/${dataset}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/output-files.txt
    echo "${target_dir}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/output-files.txt
    echo "output-file: ${target_dir}/${dataset}_${n}_ch${ch}.dat"
done

exit $code
