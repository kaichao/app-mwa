#!/bin/bash
m=$1
headers=""

# code_dir=`dirname $0`
# ${code_dir}/create_semas.sh "$1"
# code=$?
# [ $code -ne 0 ] && echo "[ERROR] while creating semaphores! " >&2 && exit $code 

# ip_list=()
pattern="^([^~]+)~([^/]+)/p([0-9]+)"
if [[ $m =~ $pattern ]]; then
    p=${BASH_REMATCH[3]}
    echo "p: $p"
else
    echo "[ERROR]: Invalid message $m"
    exit 1
fi

echo "LOCAL_FITS_ROOT:$LOCAL_FITS_ROOT"
m=$m~$LOCAL_FITS_ROOT/mwa/24ch
echo $m

# echo "rsync-pull,$m" >> ${WORK_DIR}/messages.txt
pi=$(( 10#$p ))
if [ $pi -lt $POINTING_BEGIN ] || [ $pi -gt $POINTING_END ] ; then
    echo "$pi out of range: $POINTING_BEGIN to $POINTING_END. omitting." >${WORK_DIR}/custom-out.txt && exit 10
fi
order=$(( ($pi - 1) % $NUM_OF_NODES ))
node=$(printf "%04d.p419" "$order")
echo $node
echo $NODES_GROUP
scalebox task add --sink-job local-copy-unpack --to-host ${NODES_GROUP}-${node} ${m}