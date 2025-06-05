#!/usr/bin/env bash

source functions.sh
source $(dirname $0)/functions.sh

# usage: downsample all .fits file in the input directory,
# and save the result to the output directory.

# downsample command:

# psrfits_subband -dstime 4 -o 1165080856_J0630_4dt 1165080856_J0630_5s.fits 
##### only downsamples in time

# psrfits_subband -dstime 4 -outbits 4 -adjustlevels -o 1165080856_J0630_4dt 1165080856_J0630_5s.fits
##### also changes the output file to 4-bit format

# 1. set the input / output directory
if [ $INPUT_ROOT ]; then
    DIR_1CH=$(get_host_path "${INPUT_ROOT}/mwa/1ch")
else
    DIR_1CH=/cluster_data_root/mwa/1ch
fi

if [ $OUTPUT_ROOT ]; then
    DIR_1CHX=$(get_host_path "${OUTPUT_ROOT}/mwa/1chx")
else
    DIR_1CHX=/cluster_data_root/mwa/1chx
fi
# 
if [ $OUTPUT_ROOT ]; then
    DIR_1CHY=$(get_host_path "${OUTPUT_ROOT}/mwa/1chy")
else
    DIR_1CHY=/cluster_data_root/mwa/1chy
fi

# 2. check if the directory exists
m=$1
dir_1ch="$DIR_1CH/$m"
dir_1chx="$DIR_1CHX/$m"
dir_1chy="$DIR_1CHY/$m"
mkdir -p $dir_1chx; code=$?
[[ $code -ne 0 ]] && echo "[ERROR] mkdir $dir_1chx"  >> ${WORK_DIR}/custom-out.txt && exit $code

# remove existing intermediate files in output dir
rm -f $dir_1chx/*

if [ -f "${dir_1ch}/*.zst" ]; then
    # 解压文件并删除原始文件
    zstd -d --rm "${DIR_1CH}/${m}/*.zst"
fi

cd ${dir_1ch}

for f in *.fits; do
    if [ ! -e "$f" ]; then
        echo "No .fits files found in the current directory."  >> ${WORK_DIR}/custom-out.txt
        break
    fi
    echo "filename:$f"
    # 3. run the programs to downsample the files
    psrfits_subband -dstime ${DOWNSAMP_FACTOR_TIME} -o ${dir_1chx}/$f ${dir_1ch}/$f
    code=$?
    [[ $code -ne 0 ]] && echo "[ERROR] psrfits_subband, filename:${dir_1ch}/$f "  >> ${WORK_DIR}/custom-out.txt && exit $code

    # rename file to normalized
    mv ${dir_1chx}/${f}_0001.fits ${dir_1chx}/${f} && zstd --long -T2 --rm ${dir_1chx}/${f}
    code=$?
    [[ $code -ne 0 ]] && echo "[ERROR] rename fits file and zstd compress " >&2 && exit $code

    # 检查输入、输出文件的大小比例是否合理？
    post_check "${dir_1ch}/${f}" "${dir_1chx}/${f}.zst"
    code=$?
    [[ $code -ne 0 ]] && echo "[ERROR] post_check ${m} " >> ${WORK_DIR}/custom-out.txt && exit $code

    # 下采样后文件
    f0="${dir_1chx}/${f}.zst"
    # 按pointing再分发后文件
    f1="${dir_1chy}/${f}.zst"
    if [ "$ENABLE_LOCAL_COMPUTE" != "yes" ]; then
        # 非本地计算，对fits.zst文件按pointing直接做分发
        regex='^(.+)/(p[0-9]+_[0-9]+)/(t[0-9]+_[0-9]+)/(ch[0-9]+)/(p[0-9]+).fits.zst$'
        if [[ $f1 =~ $regex ]]; then
            # echo "${BASH_REMATCH[1]}"  # mwa/1chx/1257617424
            # echo "${BASH_REMATCH[2]}"  # p00001_00024
            # echo "${BASH_REMATCH[3]}"  # t1257617426_1257617505
            # echo "${BASH_REMATCH[4]}"  # ch109
            # echo "${BASH_REMATCH[5]}"  # p00001
            f1="${BASH_REMATCH[1]}/${BASH_REMATCH[5]}/${BASH_REMATCH[3]}/${BASH_REMATCH[4]}.fits.zst"
            mkdir -p "$(dirname $f1)" && mv -f $f0 $f1
            code=$?
            [[ $code -ne 0 ]] && echo "[ERROR] rename ${f0} " >> ${WORK_DIR}/custom-out.txt && exit $code

            echo $f1 > ${WORK_DIR}/output-files.txt
        else
            echo "format error, filename:$f0" >&2
            exit 81
        fi
    else
        echo $f0 >> ${WORK_DIR}/output-files.txt
    fi
    echo "${dir_1ch}/${f}" >> ${WORK_DIR}/input-files.txt
done

[ "$KEEP_SOURCE_FILE" == "no" ] && echo "${dir_1ch}" >> ${WORK_DIR}/removed-files.txt
# 用于测试
[ "$KEEP_TARGET_FILE" == "no" ] && echo "${dir_1chx}" >> ${WORK_DIR}/removed-files.txt

echo "removed-files:" >> ${WORK_DIR}/custom-out.txt
cat ${WORK_DIR}/removed-files.txt >> ${WORK_DIR}/custom-out.txt

if [ "$ENABLE_LOCAL_COMPUTE" != "yes" ]; then
    rmdir $dir_1chx
fi
echo "$1" >> ${WORK_DIR}/messages.txt

exit $code
