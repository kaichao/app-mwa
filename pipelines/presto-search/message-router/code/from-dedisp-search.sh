#!/bin/bash
# input message: 1257010784/p00001/dm1/group1

m0=$1
m=${m0%/*}
echo "new message:$m"
p=${m%/*}
echo "pointing:$p"

from_ip=$2
echo from_ip:$from_ip

sema1="fits-24ch-dedisp-completed:$p"
n1=$(scalebox semaphore countdown "$sema1")
code=$?

# Checking the exit status of the `scalebox` command
# If there is an error, print an error message and exit with the same code
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore countdown! " >&2 && exit $code 

# Checking if the semaphore is 0
if [ "$n1" -eq 0 ]; then
    # echo "clean-up,$p" >> $WORK_DIR/messages.txt
    scalebox task add --sink-job clean-up --to-ip $from_ip ${p}
fi

sema="dm-group-ready:$m"
n=$(scalebox semaphore countdown "$sema")
code=$?

# Checking the exit status of the `scalebox` command
# If there is an error, print an error message and exit with the same code
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore countdown! " >&2 && exit $code 

# Checking if the semaphore is 0
if [ "$n" -eq 0 ]; then
    # echo "fold,$m" >> $WORK_DIR/messages.txt
    scalebox task add --sink-job fold --to-ip $from_ip ${m}
fi