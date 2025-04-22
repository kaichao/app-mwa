#!/bin/bash
# input message: 1257010784/p00001/t1257010786_1257010935.fits.zst

# pwd
# ips=$(cat /app/bin/ip_list.txt)
# ip_list=($ips)

p=$1
echo $p
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
# echo "rfi-find,$p" > $WORK_DIR/messages.txt
pa=${p##*p}
pi=$(( 10#$pa ))
# order=$(( ($pi - 1) % $NUM_OF_NODES ))
# node=$(printf "%04d.p419" "$order")

# order=$(( ($pi - 1) % ${#ip_list[@]} ))
# node=${ip_list[$order]}
# echo $node
echo $NODES_GROUP
/app/bin/get_hosts.py ${pi}
code=$?
[ $code -ne 0 ] && echo "[ERROR] looking for valid hosts! " >&2 && exit 20
node=$( cat ./host.txt ) && rm ./host.txt
echo $node
# scalebox task add --sink-job rfi-find -h to_host=${NODES_GROUP}-${node} ${p}

echo source_url=$2
scalebox task add --sink-job local-copy -h to_host=${node} -h source_url=$2 ${p}