#!/bin/bash

source $(dirname $0)/functions.sh

env | sort > ${WORK_DIR}/custom-out.txt

# OBSID/BEG_END/ch/PTHEAD_PTTAIL
m=$1
# m="1257010784/1257010986_1257011185/132/00001_00003"

KEEP_SOURCE_FILE=${KEEP_SOURCE_FILE:-"yes"}

if [ $LOCAL_CAL_ROOT ]; then
    DIR_CAL="/local${LOCAL_CAL_ROOT}/mwa/cal"
else
    DIR_CAL=/data/mwa/cal
fi
if [ $LOCAL_INPUT_ROOT ]; then
    DIR_DAT="/local${LOCAL_INPUT_ROOT}/mwa/dat"
else
    DIR_DAT=/data/mwa/dat
fi
if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_1CH="/local${LOCAL_OUTPUT_ROOT}/mwa/1ch"
else
    DIR_1CH=/data/mwa/1ch
fi

my_arr=($(echo $m | tr "_" "\n" | tr "/" "\n"))
OBSID=${my_arr[0]}
BEG=${my_arr[1]}
END=${my_arr[2]}
ch=${my_arr[3]}
PTHEAD=${my_arr[4]}
PTTAIL=${my_arr[5]}
let ii=$((10#${ch}))-108
printf -v i "%02d" $ii

dat_dir="${DIR_DAT}/${OBSID}/ch${ch}/${BEG}_${END}"

# 加载UTT等元数据信息
# source ${DIR_CAL}/${OBSID}/mb_meta.env

UTT=$( /app/bin/gps2utc.py ${BEG} )
echo UTT=${UTT}
# UTT=2019-11-05T17:43:25.00
# PTLIST=${BASEDIR}/1257010784_grid_positions_f0.85_d0.3098_l102.txt
PTLIST=${DIR_CAL}/${OBSID}/pointings.txt
POINTS=$(awk "NR>=${PTHEAD} && NR<=${PTTAIL} {printf \"%s\", \$0; if (NR!=${PTTAIL}) printf \",\"}" ${PTLIST})

cd ${WORK_DIR}

make_beam -o ${OBSID} -b ${BEG} -e ${END} \
        -P ${POINTS} \
        -z ${UTT} \
        -d ${dat_dir} -f ${ch} \
        -m ${DIR_CAL}/${OBSID}/metafits_ppds.fits \
        -F ${DIR_CAL}/${OBSID}/flagged_tiles.txt \
        -J ${DIR_CAL}/${OBSID}/DI_JonesMatrices_node0${i}.dat \
        -B ${DIR_CAL}/${OBSID}/BandpassCalibration_node0${i}.dat \
        -t 6000 -W 10000 -s 
code=$?
[[ $code -ne 0 ]] && echo exit after make_beam, error_code:$code >&2 && exit $code

# 将生成的fits文件转移到规范目录下
declare -i i=0
point_arr=($(echo $POINTS | tr "," "\n" ))
for ii in $(seq $PTHEAD $PTTAIL); do
    pi=$(printf "%05d" $ii)
    dest_file_r=${OBSID}/p${pi}/t${BEG}_${END}/ch${ch}.fits
    dest_file=${DIR_1CH}/${dest_file_r}
    orig_file=${WORK_DIR}/${point_arr[${i}]}/*.fits
    mkdir -p $(dirname ${dest_file}) && mv $orig_file $dest_file
    code=$?
    [[ $code -ne 0 ]] && echo "exit after mkdir and mv, dest_file:$dest_file, error_code:$code" >&2 && exit $code

    i=$((i + 1))
done

# 检查输出文件是否完整
post_check $OBSID $ch $PTHEAD $PTTAIL $BEG $END $DIR_1CH
code=$?
[[ $code -ne 0 ]] && echo "exit after post-check output files, exit_code:$code" >&2 && exit $code

for ii in $(seq $PTHEAD $PTTAIL); do
    pi=$(printf "%05d" $ii)
    dest_file_r=${OBSID}/p${pi}/t${BEG}_${END}/ch${ch}.fits
    dest_file=${DIR_1CH}/${dest_file_r}

    # 文件压缩
    if [ "$ZSTD_TARGET_FILE" = "yes" ]; then
        zstd -T8 --rm $dest_file
        dest_file="${dest_file}.zst"
    fi

    # 输出消息 
    echo $dest_file_r >> ${WORK_DIR}/task-body.txt
    
    # scalebox task add --headers '{"sorted_tag":"1111"}' $dest_file_r; code=$?
    # [[ $code -ne 0 ]] && echo "[ERROR] send-message, msg-body:$dest_file_r, error_code:$code" >&2 && exit $code

    # 统计输出文件的字节数
    echo $dest_file >> ${WORK_DIR}/output-files.txt
done
# group-task-add 
scalebox task add; code=$?
[[ $code -ne 0 ]] && echo "[ERROR] task-add, msg-file:task-body.txt, error_code:$code" >&2 && exit $code

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
    # for ((n=BEG; n<=END; n++)); do
    #     file_name="${OBSID}/${OBSID}_${n}_ch${ch}.dat"
    #     echo "file_name to remove:${DIR_DAT}/${file_name}"
    #     rm -f "${DIR_DAT}/${file_name}"
    # done
    rm -rf ${dat_dir}
fi

exit $code
