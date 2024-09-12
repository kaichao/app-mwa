#!/bin/bash
m=$1
headers=""

# code_dir=`dirname $0`
# ${code_dir}/create_semas.sh "$1"
# code=$?
# [ $code -ne 0 ] && echo "[ERROR] while creating semaphores! " >&2 && exit $code 

# ip_list=()
# pattern="^([^~]+)~([^/]+)/p([0-9]+)"
pattern="^([^/]+)/p([0-9]+)/t([0-9]+)_([0-9]+).fits.zst"
if [[ $m =~ $pattern ]]; then
    p=${BASH_REMATCH[2]}
    echo "p: $p"
else
    echo "[ERROR]: Invalid message $m"
    exit 1
fi

# m=${m#*~}
echo $m

# echo "rsync-pull,$m" >> ${WORK_DIR}/messages.txt
pi=$(( 10#$p ))
if [ $pi -lt $POINTING_BEGIN ] || [ $pi -gt $POINTING_END ] ; then
    echo "$pi out of range: $POINTING_BEGIN to $POINTING_END. omitting." >${WORK_DIR}/custom-out.txt && exit 0
fi
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
# scalebox task add --sink-job local-copy-unpack --to-host ${NODES_GROUP}-${node} ${m}
scalebox task add --sink-job local-copy ${m}
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt