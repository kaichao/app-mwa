#!/bin/bash

# source functions.sh

#1257010784/p00001
m=$1
headers=$2

# set the file paths. We'll need this later.
if [ $PLAN_FILE ]; then
    DDPLAN_FILE=/app/bin/$PLAN_FILE
else
    echo "[ERROR] Search plan not set!" >&2 && exit 10
fi

pattern='"source_url":"([^"]+)"'
if [[ $headers =~ $pattern ]]; then
    source_url="${BASH_REMATCH[1]}"
    echo "source_url: $source_url"
else
    # no from_job in json 
    source_url=""
fi

# parse the input message
dataset="${m%%/*}"
pointing="${m#*/}"

echo "dataset: $dataset, pointing: $pointing"

echo "total lines: ${MAX_LINENUM}"
file_path=${DDPLAN_FILE}
if [ -f $file_path ]; then
    total_lines=$(wc -l < $file_path)
    date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
    
    
# the first line is the header, skip it
    for ((i=2; i<=$total_lines; i++)); do
        line=$(sed -n "${i}p" $file_path)
        calls=$(echo $line | awk '{print $8}')
        NCALLS=$(echo $line | awk '{print $9}')
        j=$((i - 1))
        # echo "line $j: $calls"
        dmi=$(printf "%02d" "$j")
        sema="dm-group-ready:$dataset/$pointing/dm$dmi"
        # echo "$sema"
        scalebox semaphore create $sema $(($calls/$NCALLS))
    done
    
    sema2="pointing-finished:$dataset/$pointing"
    # echo "$sema2"
    scalebox semaphore create $sema2 $NUM_GROUPS
    date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
    
    echo "source_url: $source_url" > ${WORK_DIR}/custom-out.txt
    # for different sources, send message to different tasks
    # if the source is an ip, send to local-copy-unpack
    if [[ $source_url =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        scalebox variable set local_pointing:$dataset/$pointing yes
        scalebox task add --sink-job=local-copy-unpack --to-ip=$source_url $dataset/$pointing
    else
        scalebox variable set local_pointing:$dataset/$pointing no
        scalebox task add --sink-job=local-wait-queue -h source_url=$source_url $dataset/$pointing
    fi
else
    echo "DDplan file not found: $file_path"
    exit 2
fi
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

