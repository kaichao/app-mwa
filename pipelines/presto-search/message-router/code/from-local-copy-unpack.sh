#!/bin/bash
# input message: 1257010784/p00001/t1257010786_1257010935.fits.zst
m=$1

echo $m

p=${m%/*}
sema="fits-24ch-presto-ready:$p"
n=$(scalebox semaphore countdown "$sema")
code=$?

from_ip=$2
# Checking the exit status of the `scalebox` command
# If there is an error, print an error message and exit with the same code
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore countdown! " >&2 && exit $code 

# Checking if the semaphore is 0
if [ "$n" -eq 0 ]; then
    echo $p
    # echo "rfi-find,$p" > $WORK_DIR/messages.txt
    scalebox task add --sink-job rfi-find --to-ip $from_ip ${p}
fi