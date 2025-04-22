#!/bin/bash
# input message: 1257010784/p00001
m=$1
echo $m

from_ip=$2
echo $from_ip
echo "total lines: ${MAX_LINENUM}"
echo "total groups:" ${NUM_GROUPS}

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

# scalebox task add --sink-job dedisp-search -h to_ip=$from_ip ${m}
for i in $( seq 1 ${NUM_GROUPS} )
do
    j=$( printf "%03d" "$i" )
    echo ${m}/${j} >> ${WORK_DIR}/task-body.txt
done

scalebox task add --sink-job dedisp-search -h to_ip=$from_ip --task-file ${WORK_DIR}/task-body.txt


date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt