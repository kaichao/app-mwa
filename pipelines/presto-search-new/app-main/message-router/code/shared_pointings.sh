#!/bin/bash

#1257010784/p00001_00048
m=$1

# set the file paths. We'll need this later.
if [ $PLAN_FILE ]; then
    DDPLAN_FILE=/app/bin/$PLAN_FILE
else
    echo "[ERROR] Search plan not set!" >&2 && exit 10
fi

# parse the input message
dataset="${m%%/*}"
prange="${m#*/p}"

echo $prange
PBS="${prange%%_*}"
PES="${prange#*_}" 

echo "$PBS, $PES"
PB=$((10#$PBS))
PE=$((10#$PES))

echo "dataset: $dataset, pointing range: $PB-$PE"
# m1=${m%~*}
# dataset=`echo $m | awk -F "~" '{print $2}'`
echo "total lines: ${MAX_LINENUM}"
file_path=${DDPLAN_FILE}
if [ -f $file_path ]; then
    total_lines=$(wc -l < $file_path)
    date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
    /app/bin/list_missing.sh $dataset pointings.txt $PB $PE

for pointing in $( cat ${WORK_DIR}/pointings.txt ); do
        # pi=$(printf "%05d" "$p")

    # the first line is the header, skip it
        for ((i=2; i<=$total_lines; i++)); do
            line=$(sed -n "${i}p" $file_path)
            calls=$(echo $line | awk '{print $8}')
            NCALLS=$(echo $line | awk '{print $9}')
            j=$((i - 1))
            # echo "line $j: $calls"
            sema="dm-group-ready:$dataset/$pointing/dm$j"
            # echo "$sema"
            scalebox semaphore create $sema $(($calls/$NCALLS))
        done
        
        sema2="pointing-finished:$dataset/$pointing"
        echo "$sema2"
        scalebox semaphore create $sema2 $NUM_GROUPS
        date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
        scalebox task add --sink-job=local-wait-queue $dataset/$pointing
    done
else
    echo "DDplan file not found: $file_path"
    exit 2
fi
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

