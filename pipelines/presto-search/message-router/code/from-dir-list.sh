#!/bin/bash
m=$1
headers=""

# code_dir=`dirname $0`
# ${code_dir}/create_semas.sh "$1"
# code=$?
# [ $code -ne 0 ] && echo "[ERROR] while creating semaphores! " >&2 && exit $code 

ip_list=("10.11.16.79" "10.11.16.76" "10.11.16.75")
pattern="^([^~]+)~([0-9]+)/p([0-9]+)"
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
echo $pi
order=$(( ($pi - 1) % ${#ip_list[@]} ))
echo $order
scalebox task add --sink-job local-copy-unpack --to-ip ${ip_list[$order]} ${m}