#!/bin/bash
# input message: 1257010784/p00001/dm1/group1

m0=$1
m=${m0%/*}
echo "new message:$m"
p=${m%/*}
echo "pointing:$p"
dm=${m##*/} # dm1
# echo "dm:$dm"

from_ip=$2
echo from_ip:$from_ip

# in case we are testing, do not remove the raw files.
sema1="pointing-finished:$p"
echo $sema1
n1=$(scalebox semaphore decrement "$sema1")
code=$?
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore decrement! " >&2 && exit $code 

date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

host=$(/app/bin/get_hostname.py $from_ip)
host=${host%%.*}
echo $host

echo $n1
# Checking if the semaphore is 0
if [ "$n1" -eq 0 ]; then
    # echo "clean-up,$p" >> $WORK_DIR/messages.txt
    # echo ${SHARED_ROOT}/mwa/24ch/${p} >&2
    echo ${LOCAL_FITS_ROOT}/mwa/24ch/${p} >&2
    # ssh -p ${SSH_PORT} ${DEFAULT_USER}@${from_ip} rm -rf ${LOCAL_FITS_ROOT}/mwa/24ch/${p} ${SHARED_ROOT}/mwa/24ch/${p} ${LOCAL_SHM_ROOT}/mwa/dedisp/${p}/RFIfile*
    ssh -p ${SSH_PORT} ${DEFAULT_USER}@${from_ip} rm -rf ${LOCAL_FITS_ROOT}/mwa/24ch/${p}

    check the waiting tasks from db
    check_result=$(python /app/bin/check_tasks.py ${from_ip})
    code=$?
    [ $code -ne 0 ] && echo "[ERROR] in check_task_queue.py! " >&2 && exit $code
    echo "The check_tasks result: $check_result"
    if [ "$check_result" -eq 0 ]; then
        # we can send a message to redis server.
        # the message is in the format of "host:timestamp", with priority $n3
        redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD $REDIS_QUEUE 1 "$from_ip:$(date +%s%3N)"
    else
        sema3="host_vtask_size:local-copy:$host"
        n3=$(scalebox semaphore increment $sema3)
        code=$?
        [ $code -ne 0 ] && echo "[ERROR] scalebox semaphore $sema3 increment! " >&2 && exit $code 
        
        sema2="global_vtask_size:local-wait-queue"
        n2=$(scalebox semaphore increment $sema2)
        code=$?
        [ $code -ne 0 ] && echo "[ERROR] scalebox semaphore $sema2 increment! " >&2 && exit $code
    # fi
fi
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
sema="dm-group-ready:$m"
n=$(scalebox semaphore decrement "$sema")
code=$?

# Checking the exit status of the `scalebox` command
# If there is an error, print an error message and exit with the same code
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore decrement! " >&2 && exit $code 
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt
# Checking if the semaphore is 0
if [ "$n" -eq 0 ]; then
    # echo "fold,$m" >> $WORK_DIR/messages.txt
    echo dm:$dm >${WORK_DIR}/custom-out.txt
    if [ "$dm" == "dm09" ]; then
        sema4="host_vtask_size:local-copy-unpack:$host"
        n4=$(scalebox semaphore increment $sema4)
        echo $sema4
        code=$?
        [ $code -ne 0 ] && echo "[ERROR] scalebox semaphore $sema4 increment! " >&2 && exit $code 
    fi
    scalebox task add --sink-job fold -h to_ip=$from_ip ${m}
fi