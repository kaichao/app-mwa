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

# target_url is local-dir
target_url=$(get_header "$2" "target_url")
target_subdir=$(get_header "$2" "target_subdir")
if [[ "$target_url" == /* ]]; then
    target_dir="${target_url}/dat/${target_subdir}"
else
    target_dir="${LOCAL_TMPDIR}/${target_url}/dat/${target_subdir}"
fi
mkdir -p "${target_dir}"

source_url=$(get_header "$2" "source_url")
source_mode=$(get_mode "$source_url")
source_dir=$(get_data_root "$source_url")

global_dat_dir=$(get_header "$2" "_global_dat_dir")
# 设置--touch，将文件更新时间更新为当前时间，以免在/tmp中被删除
if [ -n "$global_dat_dir" ]; then
    # 容器内目录与本机目录映射，加上/local_data_root
    global_dat_dir=$(get_host_path $global_dat_dir)
    global_dat_dir="${global_dat_dir}/dat/${target_subdir}"
    # 单路读入，双路输出
    cmd_suffix="tee >(tar -C ${target_dir} --touch -xvf -) >(tar -C ${global_dat_dir} --touch -xvf -) > /dev/null"
    mkdir -p "${global_dat_dir}"
else
    cmd_suffix="tar -C ${target_dir} --touch -xvf -"
fi

# BW_LIMIT  "500k"/"1m"
bw_limit=$(get_header "$2" "bw_limit")

if [ "$source_mode" = "LOCAL" ]; then
    source_dir=$(get_host_path $source_dir)
    source_file="${source_dir}/tar/$dataset/$file_name"

    if [ -n "$bw_limit" ]; then
        # 已设置bw_limit
        cmd_prefix="pv -q -L ${bw_limit} < ${source_file} | zstd -d"
    else
        cmd_prefix="zstd -dc ${source_file}"
    fi
else
    # SSH加载
    ssh_cmd=$(get_ssh_cmd "$2" "source_url" "source_jump")
    source_file="${source_dir}/tar/$dataset/$file_name"
    cmd_prefix="zstd -d"
    if [ -n "$bw_limit" ]; then
        # 已设置bw_limit
        cmd_prefix="pv -q -L ${bw_limit}| ${cmd_prefix}"
    fi
    
    # 安全构建SSH命令：使用printf %q转义参数，避免引号丢失问题
    # 使用printf %q转义文件路径，确保特殊字符被正确处理
    printf -v escaped_file_path '%q' "${source_dir}/tar/$dataset/$file_name"
    
    # 构建远程命令 - 使用转义后的路径
    remote_cmd="cat $escaped_file_path -"
    
    # 构建完整的SSH命令
    # 注意：这里假设ssh_cmd已经是一个有效的ssh命令字符串
    # 例如：ssh -p 22 user@host
    cmd_prefix="$ssh_cmd \"$remote_cmd\" | ${cmd_prefix}"
fi

cmd="${cmd_prefix} | ${cmd_suffix}"

echo "cmd_prefix:$cmd_prefix" >> ${WORK_DIR}/auxout.txt
echo "cmd_suffix:$cmd_suffix" >> ${WORK_DIR}/auxout.txt
echo "cmd:$cmd" >> ${WORK_DIR}/auxout.txt

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

echo "source_url:$source_url, source_dir:$source_dir" >> ${WORK_DIR}/auxout.txt
echo "target_url:$target_url, target_dir:$target_dir, target_subdir:$target_subdir" >> ${WORK_DIR}/auxout.txt
echo "task-body:$1" >> ${WORK_DIR}/auxout.txt

cd "${target_dir}" && eval "$cmd"
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]exit after pull-unpack, error_code:$code"  >> ${WORK_DIR}/auxout.txt && exit $code

chmod 644 "${target_dir}/*.dat"
if [ -n "$global_dat_dir" ]; then
    chmod 644 "${global_dat_dir}/*.dat"
fi

# 检查输出文件是否完整
post_check $dataset $ch $begin $end $target_dir
code=$?
[[ $code -ne 0 ]] && echo "[ERROR]exit after post-check output files, exit_code:$code"  >> ${WORK_DIR}/auxout.txt && exit $code

for ((n=$begin; n<=$end; n++))
do
    echo "${target_dir}/${dataset}_${n}_ch${ch}.dat" >> ${WORK_DIR}/output-files.txt
done

echo "$1" >> ${WORK_DIR}/sink-tasks.txt

if [ "$source_mode" = "LOCAL" ]; then
    echo "${source_file}" >> ${WORK_DIR}/input-files.txt
    keep_source_file=$(get_header "$2" "keep_source_file")
    if [ "$keep_source_file" = "no" ]; then
        echo "${source_file}" > ${WORK_DIR}/removed-files.txt
    fi
fi

exit $code
