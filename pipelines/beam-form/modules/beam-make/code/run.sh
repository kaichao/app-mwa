#!/usr/bin/env bash

source functions.sh
source $(dirname $0)/functions.sh

env | sort > ${WORK_DIR}/custom-out.txt

# OBSID/p{PTHEAD}_{PTTAIL}/t{BEG}_{END}/ch{ch}/
# m="1257617424/p00001_00048/t1257617426_1257617505/ch109"
m=$1
pointing_range=$(get_parameter "$2" "pointing_range")

KEEP_SOURCE_FILE=${KEEP_SOURCE_FILE:-"yes"}

if [ $LOCAL_CAL_ROOT ]; then
    DIR_CAL="/local${LOCAL_CAL_ROOT}/mwa/cal"
else
    DIR_CAL=/cluster_data_root/mwa/cal
fi
if [ $LOCAL_INPUT_ROOT ]; then
    DIR_DAT="/local${LOCAL_INPUT_ROOT}/mwa/dat"
else
    DIR_DAT=/cluster_data_root/mwa/dat
fi
if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_1CH="/local${LOCAL_OUTPUT_ROOT}/mwa/1ch"
else
    DIR_1CH=/cluster_data_root/mwa/1ch
fi

my_arr=($(echo $m | tr "_" "\n" | tr "/" "\n"))
OBSID=${my_arr[0]}
PTHEAD=${my_arr[1]:1}
PTTAIL=${my_arr[2]}
BEG=${my_arr[3]:1}
END=${my_arr[4]}
ch=${my_arr[5]:2}

ii=$(( ch - 108 ))
# let ii=$((10#${ch}))-108
printf -v i "%02d" $ii

dat_dir="${DIR_DAT}/${OBSID}/${pointing_range}/t${BEG}_${END}/ch${ch}"
#dat_dir="${DIR_DAT}/${OBSID}/p${PTHEAD}_${PTTAIL}/t${BEG}_${END}/ch${ch}"

UTT=$( /app/bin/gps2utc.py ${BEG} )
# UTT=2019-11-05T17:43:25.00

echo UTT=${UTT}
echo "dat_dir=${dat_dir}"

# PTLIST=${BASEDIR}/1257010784_grid_positions_f0.85_d0.3098_l102.txt
PTLIST=${DIR_CAL}/${OBSID}/pointings.txt
POINTS=$(awk "NR>=${PTHEAD} && NR<=${PTTAIL} {printf \"%s\", \$0; if (NR!=${PTTAIL}) printf \",\"}" ${PTLIST})
echo "POINTS:${POINTS}"

cd ${WORK_DIR}
if [ "$RUNNING_MODE" = "1" ]; then
    make_beam -o ${OBSID} -b ${BEG} -e ${END} \
        -P ${POINTS} \
        -z ${UTT} \
        -d ${dat_dir} -f ${ch} \
        -m ${DIR_CAL}/${OBSID}/metafits_ppds.fits \
        -F ${DIR_CAL}/${OBSID}/flagged_tiles.txt \
        -t 6000 -W 10000 -s \
        -0 ${DIR_CAL}/${OBSID}/calibration_solution.bin -c 23
else
    make_beam -o ${OBSID} -b ${BEG} -e ${END} \
        -P ${POINTS} \
        -z ${UTT} \
        -d ${dat_dir} -f ${ch} \
        -m ${DIR_CAL}/${OBSID}/metafits_ppds.fits \
        -F ${DIR_CAL}/${OBSID}/flagged_tiles.txt \
        -t 6000 -W 10000 -s \
        -J ${DIR_CAL}/${OBSID}/DI_JonesMatrices_node0${i}.dat \
        -B ${DIR_CAL}/${OBSID}/BandpassCalibration_node0${i}.dat
fi

code=$?
[[ $code -ne 0 ]] && echo exit after make_beam, error_code:$code >&2 && exit $code

fits_dir=${DIR_1CH}/${OBSID}/p${PTHEAD}_${PTTAIL}/t${BEG}_${END}/ch${ch}
# 将生成的fits文件转移到规范目录下
declare -i i=0
point_arr=($(echo $POINTS | tr "," "\n" ))
for ii in $(seq $PTHEAD $PTTAIL); do
    pi=$(printf "%05d" $ii)
    # dest_file_r=${OBSID}/p${PTHEAD}_${PTTAIL}/t${BEG}_${END}/ch${ch}/p${pi}.fits
    # dest_file=${DIR_1CH}/${dest_file_r}
    dest_file=${fits_dir}/p${pi}.fits
    orig_file=${WORK_DIR}/${point_arr[${i}]}/*.fits

    # BUG：压缩参数开启会导致post_check检查出错 ！
    if [ "$ZSTD_TARGET_FILE" = "yes" ]; then
        zstd -T8 --rm ${orig_file}
        orig_file="${orig_file}.zst"
        dest_file="${dest_file}.zst"
    fi

    mkdir -p ${fits_dir} && mv -f $orig_file $dest_file
    code=$?
    [[ $code -ne 0 ]] && echo "exit after mkdir and mv, dest_file:$dest_file, error_code:$code" >&2 && exit $code

    # 统计输出文件的字节数
    echo $dest_file >> ${WORK_DIR}/output-files.txt

    i=$((i + 1))
done

# 检查输出文件是否完整
post_check $OBSID $ch $PTHEAD $PTTAIL $BEG $END $DIR_1CH
code=$?
[[ $code -ne 0 ]] && echo "exit after post-check output files, exit_code:$code" >&2 && exit $code

echo $1 > ${WORK_DIR}/messages.txt

# scalebox task add "$1"; code=$?
# [[ $code -ne 0 ]] && echo "[ERROR] task-add, error_code:$code" >&2 && exit $code

# 统计输入文件的总字节数
num_points=${#point_arr[@]}
num_files=$(expr "$END" - "$BEG")
file_length=327680000
input_bytes=$(( (num_files+1) * file_length * num_points ))
echo '{
    "inputBytes":'${input_bytes}'
}' > ${WORK_DIR}/task-exec.json

if [ "$KEEP_SOURCE_FILE" = "no" ]; then
    # only used for test
    echo "remove dat files"
    echo ${dat_dir} >> ${WORK_DIR}/removed-files.txt
fi
if [ "$KEEP_TARGET_FILE" = "no" ]; then
    # only used for test
    echo "remove fits files"
    echo ${fits_dir} >> ${WORK_DIR}/removed-files.txt
fi

echo stdout,code=$code 
echo stderr >&2

exit $code
