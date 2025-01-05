#!/bin/bash

# 检查输出文件是否有效？
function post_check() {
    dataset=$1
    ch=$2
    p0=$3
    p1=$4
    t0=$5
    t1=$6
    target_dir=$7

    echo "[post_check]:" >> ${WORK_DIR}/custom-out.txt
    echo "ch=$ch"  >> ${WORK_DIR}/custom-out.txt
    echo "p0=$p0, p1=$p1"  >> ${WORK_DIR}/custom-out.txt
    echo "t0=$t0, t1=$t1"  >> ${WORK_DIR}/custom-out.txt
    echo "target_dir=$target_dir"  >> ${WORK_DIR}/custom-out.txt

    # 初始化字节数数组
    sizes=()
    for ii in $(seq $p0 $p1); do
        pi=$(printf "%05d" $ii)
        dest_file_r="${dataset}/p${p0}_${p1}/t${t0}_${t1}/ch${ch}//p${pi}.fits"
        dest_file=${target_dir}/${dest_file_r}

        if [[ -f "$dest_file" ]]; then
            sizes+=( $(stat -c%s "$dest_file") )
            echo "file: $dest_file_r, bytes: $(stat -c%s "$dest_file")" >> ${WORK_DIR}/custom-out.txt
        else
            echo "[ERROR] post_check file $dest_file not exists!" >> ${WORK_DIR}/custom-out.txt
            exit 101
        fi
    done

    # 获取文件数量
    num_files=${#sizes[@]}
    # 计算均值
    mean=0
    for size in "${sizes[@]}"; do
        mean=$((mean + size))
    done
    mean=$((mean / num_files))
    echo " $num_files files, average bytes: $mean" >> ${WORK_DIR}/custom-out.txt

    # 检查文件字节数是否一致？
    all_equal=true
    for (( i=0; i<num_files-1; i++ )); do
        if [[ "${sizes[i]}" -ne "${sizes[0]}" ]]; then
            all_equal=false
            break
        fi
    done

    if $all_equal; then
        echo "所有文件字节数一致" >> ${WORK_DIR}/custom-out.txt
    else
        echo [ERROR] "文件字节数不一致" >> ${WORK_DIR}/custom-out.txt
        return 102
    fi
}
