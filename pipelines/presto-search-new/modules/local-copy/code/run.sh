#!/bin/bash

# similar to pull-unpack
# copy the whole directory to local target directory
# input format: dataset/pointing
# get other info from headers

source functions.sh
source /app/share/bin/functions.sh

if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_FITS="/local_data_root${LOCAL_OUTPUT_ROOT}/mwa/24ch"
else
    DIR_FITS=/cluster_data_root/mwa/24ch
fi

m0=$1
headers=$2

# check the input message format using regex: should be like 123123123/p01235

pattern="^([^/]+)/(p[0-9]+)$"
echo $m0
echo $pattern
if [[ $m0 =~ $pattern ]]; then
    m=$m0
    dataset=${BASH_REMATCH[1]}
    pointing=${BASH_REMATCH[2]}
else
    echo "[ERROR] Invalid message format: $m0" >&2
    exit 1
fi

source_url=$(get_header "$2" "source_url")
source_mode=$(get_mode "$source_url")
source_dir=$(get_data_root "$source_url")

# if the mode is LOCAL, set the variable $INPUT_ROOT
if [ "$source_mode" == "LOCAL" ]; then
    INPUT_ROOT="/local_data_root${source_dir}"
else
    echo $source_mode
    INPUT_ROOT="${source_dir}"
fi

target_dir="$DIR_FITS/${m}"
echo "file dir:${target_dir}"

input_dir="$INPUT_ROOT/${m}"

# if the mode is local, copy the directory to local target directory with given bandwidth
# the target dir maybe existing, so make the dir and copy each file to the target dir
# try to use pv to set the bandwidth
if [ "$source_mode" == "LOCAL" ]; then
    # first check if the input dir exists
    if [ ! -d "${input_dir}" ]; then
        echo "[ERROR] Input directory ${input_dir} does not exist" >&2 && exit 11
    fi
    mkdir -p ${target_dir}
    # if $BW_LIMIT is set, generate a new variable to use in pv; else set it to empty string
    if [ -n "$BW_LIMIT" ]; then
        BW_LIMIT_ARG="-L ${BW_LIMIT}"
    else
        BW_LIMIT_ARG=""
    fi

    echo BW_LIMIT_ARG:$BW_LIMIT_ARG
    echo "Copying ${input_dir} to ${target_dir}" >${WORK_DIR}/custom-out.txt
    for file in $( ls ${input_dir} ); do
        echo "Copying ${file} to ${target_dir}"
        pv -q $BW_LIMIT_ARG ${input_dir}/${file} > ${target_dir}/${file}
    done
else
    echo "Copying from remote server..." >${WORK_DIR}/custom-out.txt
    mkdir -p ${target_dir}
    # if the mode is remote, copy the directory to local target directory with given bandwidth
    # we will use rsync over ssh to copy the directory    
    source_ssh_option=$(get_ssh_option "$2" "source_url" "jump_servers")
    echo $source_ssh_option
    # the source_ssh_option is like -p 10022 -J XXXXXXXXXX, so we need to add -e ssh to the rsync command
    rsync_option=""
    # if $RSYNC_BW_LIMIT is set, add it to the rsync_option
    if [ -n "$RSYNC_BW_LIMIT" ]; then
        rsync_option="${rsync_option} --bwlimit=${RSYNC_BW_LIMIT}"
    fi
    # if $ZSTD_LEVEL is set, add it to the rsync_option
    if [ -n "$ZSTD_LEVEL" ]; then
        rsync_option="${rsync_option} --compress --compress-choice=zstd --compress-level=${ZSTD_LEVEL}"
    fi
    echo $rsync_option
    source_ssh_url=$(to_ssh_url $source_url)
    # now generate the rsync command and print it
    cmd="rsync -avz $rsync_option -e \"ssh ${source_ssh_option}\" ${source_ssh_url}/${m}/ ${target_dir} "
    if [ -n "$TASK_TIMEOUT_SECONDS" ]; then
            # 若timeout超时，返回124编码。否则实际完成后，返回0。导致后续脚本错误，误删除出错的源文件
            cmd="timeout ${TASK_TIMEOUT_SECONDS}s $cmd"
    fi
    echo $cmd
    # execute the rsync command
    eval $cmd
    # check the return code
    code=$?
    if [ $code -ne 0 ]; then
        echo "[ERROR] Failed to copy ${input_dir} to ${target_dir}" >&2 && exit $code
    fi
fi


ls -l ${target_dir} >> ${WORK_DIR}/custom-out.txt

# send message to next module
echo ${m} >> ${WORK_DIR}/messages.txt
