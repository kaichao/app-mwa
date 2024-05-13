#!/bin/bash
# input message: 1257010784/p00001/dm1/group1
m0=$1

echo $m0

p0=${m0%/*}
p=${p0%/*}
sema="fits-24ch-dedisp-completed:$p"
n=$(scalebox semaphore countdown "$sema")
code=$?

from_ip=$2
echo $from_ip

# Checking the exit status of the `scalebox` command
# If there is an error, print an error message and exit with the same code
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore countdown! " >&2 && exit $code 

# Checking if the semaphore is 0
if [ "$n" -eq 0 ]; then
    # echo "clean-up,$p" >> $WORK_DIR/messages.txt
    scalebox task add --sink-job clean-up --to-ip $from_ip ${m}
fi
# m="root@${from_ip}${LOCAL_SHM_ROOT}/mwa/dedisp~${m0}~${LOCAL_SHM_ROOT}/mwa/dedisp/tar"
# echo "local-copy,$m" >> $WORK_DIR/messages.txt
scalebox task add --sink-job search --to-ip $from_ip ${m}