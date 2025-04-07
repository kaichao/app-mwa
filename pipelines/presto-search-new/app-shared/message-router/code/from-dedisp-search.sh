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
# sema1="pointing-finished:$p"
# n1=$(scalebox semaphore countdown "$sema1")
# code=$?
# [ $code -ne 0 ] && echo "[ERROR] scalebox semaphore countdown! " >&2 && exit $code 
# date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

# # Checking if the semaphore is 0
# if [ "$n1" -eq 0 ]; then
#     # echo "clean-up,$p" >> $WORK_DIR/messages.txt
#     echo ${SHARED_ROOT}/mwa/24ch/${p} >&2
#     echo ${LOCAL_FITS_ROOT}/mwa/24ch/${p} >&2
#     # ssh -p ${SSH_PORT} ${DEFAULT_USER}@${from_ip} rm -rf ${LOCAL_FITS_ROOT}/mwa/24ch/${p} ${SHARED_ROOT}/mwa/24ch/${p} ${LOCAL_SHM_ROOT}/mwa/dedisp/${p}/RFIfile*
#     ssh -p ${SSH_PORT} ${DEFAULT_USER}@${from_ip} rm -rf ${LOCAL_FITS_ROOT}/mwa/24ch/${p} ${LOCAL_SHM_ROOT}/mwa/dedisp/${p}/RFIfile*

# # fi
# date --iso-8601=ns >> ${WORK_DIR}/timestamps.txt

# echo "fold,$m" >> $WORK_DIR/messages.txt
scalebox task add --sink-job fold ${m}