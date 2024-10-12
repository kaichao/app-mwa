#!/bin/bash
# input message: 1257010784/p00001/t1257010786_1257010935.fits.zst

pwd
ips=$(cat /app/bin/ip_list.txt)
ip_list=($ips)

m=$1

echo $m
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
p=${m%/*}
sema="fits-24ch-presto-ready:$p"
n=$(scalebox semaphore countdown "$sema")
code=$?

# from_ip=$2
# Checking the exit status of the `scalebox` command
# If there is an error, print an error message and exit with the same code
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore countdown! " >&2 && exit $code 
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
# Checking if the semaphore is 0
if [ "$n" -eq 0 ]; then
    echo $p
    # echo "rfi-find,$p" > $WORK_DIR/messages.txt
    pa=${p##*p}
    pi=$(( 10#$pa ))
    # order=$(( ($pi - 1) % $NUM_OF_NODES ))
    # node=$(printf "%04d.p419" "$order")

    order=$(( ($pi - 1) % ${#ip_list[@]} ))
    node=${ip_list[$order]}
    echo $node
    echo $NODES_GROUP

    # scalebox task add --sink-job rfi-find --to-host ${NODES_GROUP}-${node} ${p}
    scalebox task add --sink-job rfi-find --to-host ${node} ${p}
fi