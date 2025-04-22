#!/bin/bash
# input message: 1257010784/p00001/dm1.tar.zst
m0=$1

echo $m0

from_ip=$2
echo $from_ip

# m="root@${from_ip}${LOCAL_SHM_ROOT}/mwa/dedisp~${m0}~${LOCAL_SHM_ROOT}/mwa/dedisp/tar"
# echo "local-copy,$m" >> $WORK_DIR/messages.txt

# m="${LOCAL_SHM_ROOT}/mwa/png~${m0}~${RESULT_DIR}"
m=$m0

scalebox task add --sink-job result-push -h to_ip=$from_ip ${m}