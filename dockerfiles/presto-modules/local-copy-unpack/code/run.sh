#!/bin/bash

# similar to pull-unpack
# input example: user@url/ssh/dir~1257010784/p00001/t1257010786_1257010936.fits.zst~/target/dir
# input format: dataset/pointing/filename.fits.zst

m0=$1

pattern1="(^[^/]+)([^~]+)~([^~]+)~(.+)$"
if [[ $m0 =~ $pattern1 ]]; then
    ssh_host=${BASH_REMATCH[1]}
    source_dir=${BASH_REMATCH[2]}
    m=${BASH_REMATCH[3]}
    target_dir=${BASH_REMATCH[4]}
else
    echo "[ERROR] Input does not match :$1" >&2 && exit 5
fi

dirnm=$(dirname $m)
filenm=$(basename $m)
fitsnm=${filenm%.*}

pattern2="^([0-9]+)/p([0-9]+)/t([0-9]+)_([0-9]+)\.fits\.zst$"
if [[ $m =~ $pattern2 ]]; then
    dataset=${BASH_REMATCH[1]}
    pointing=${BASH_REMATCH[2]}
    tbegin=${BASH_REMATCH[3]}
    tend=${BASH_REMATCH[4]}
else
    echo "[ERROR] Input does not match :$m" >&2 && exit 5
fi

# record important parameters
echo "dataset:$dataset" >> /work/custom-out.txt
echo "pointing:$pointing" >> /work/custom-out.txt
echo "filename:t$tbegin_$tend" >> /work/custom-out.txt

# jump_servers=$(get_parameter "$2" "jump_servers")
jump_servers=${JUMP_SERVERS}
jump_servers_option=""
if [ $jump_servers ]; then
    jump_servers_option="-J '${jump_servers}' "
fi

ssh_args="-T -c aes128-gcm@openssh.com -o Compression=no -x ${jump_servers_option}"

# source_url=$(get_parameter "$2" "source_url")
source_url="${ssh_host}:${source_dir}"
# target_url=$(get_parameter "$2" "target_url")
target_url="${target_dir}/${dirnm}"
# ssh_port=$(get_parameter "$2" "ssh_port")
ssh_port=22

# IFS=':' read -r ssh_host source_dir <<< ${source_url}

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

# echo "source_url:$source_url" >> /work/custom-out.txt
echo "source_dir:$source_dir" >> /work/custom-out.txt
echo "target_url:$target_url" >> /work/custom-out.txt
echo "ssh_host:$ssh_host" >> /work/custom-out.txt
echo "ssh_port:$ssh_port" >> /work/custom-out.txt
echo "ssh_args:$ssh_args" >> /work/custom-out.txt
echo "message:$m0" >> /work/custom-out.txt
target_dir="/local${target_url}"

cmd="ssh -p ${ssh_port} ${ssh_args} ${ssh_host} \"cat ${source_dir}/$m\" - | zstd -d > $fitsnm"
echo "cmd:$cmd"

mkdir -p ${target_dir} \
    && cd ${target_dir} \
    && eval $cmd
code=$?

[[ $code -ne 0 ]] && echo "exit after local-copy-unpack, error_code:$code" >&2 && exit $code


echo "${m}" >> ${WORK_DIR}/messages.txt
echo "${target_dir}/${filenm}" >> ${WORK_DIR}/output-files.txt
echo "output-file: ${target_dir}/${filenm}"