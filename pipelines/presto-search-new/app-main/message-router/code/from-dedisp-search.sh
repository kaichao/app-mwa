#!/bin/bash
# input message: 1257010784/p00001/dm1/group1

m0=$1
m=${m0%/*}
echo "new message:$m"
p=${m%/*}
echo "pointing:$p"

from_ip=$2
echo from_ip:$from_ip

# in case we are testing, do not remove the raw files.
sema1="pointing-finished:$p"
n1=$(scalebox semaphore decrement "$sema1")
code=$?
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore decrement! " >&2 && exit $code 
date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

# Checking if the semaphore is 0
if [ "$n1" -eq 0 ]; then
    # echo "clean-up,$p" >> $WORK_DIR/messages.txt
    # echo ${SHARED_ROOT}/mwa/24ch/${p} >&2
    echo ${LOCAL_FITS_ROOT}/mwa/24ch/${p} >&2
    # ssh -p ${SSH_PORT} ${DEFAULT_USER}@${from_ip} rm -rf ${LOCAL_FITS_ROOT}/mwa/24ch/${p} ${SHARED_ROOT}/mwa/24ch/${p} ${LOCAL_SHM_ROOT}/mwa/dedisp/${p}/RFIfile*
    ssh -p ${SSH_PORT} ${DEFAULT_USER}@${from_ip} rm -rf ${LOCAL_FITS_ROOT}/mwa/24ch/${p}

    host=$(get_hostname.py $from_ip)

    sema3="host-spare:$host"
    n3=$(scalebox semaphore increment $sema3)
    code=$?
    [ $code -ne 0 ] && echo "[ERROR] scalebox semaphore increment! " >&2 && exit $code 

    run_cached_pointings=$(scalebox variable get run_cached_pointings)
    if [ run_cached_pointings = 'no' ]; then
        # check the n3 value
        # if n3 > 0, then we can send a message to redis server.
        if [ $n3 -gt 0 ]; then
            # the message is in the format of "host:timestamp", with priority $n3
            redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD ${REDIS_KEY} $n3 "$host:$(date +%s)"
        fi

    elif [ run_cached_pointings = 'yes' ]; then
        sema2="global-vtask-size_local-wait-queue"
        n2=$(scalebox semaphore increment $sema2)
        code=$?
        [ $code -ne 0 ] && echo "[ERROR] scalebox semaphore increment! " >&2 && exit $code
    fi


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
    scalebox task add --sink-job fold --to-ip $from_ip ${m}
fi