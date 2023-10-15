#!/bin/bash

# export SERVER_INNER_MAIN=10.255.128.2
export SERVER_MAIN=159.226.237.135
export SERVER_SLAVE=60.245.209.223

rm -f *.tmp-txt

PGHOST=${SERVER_SLAVE} GRPC_SERVER=${SERVER_SLAVE} scalebox app create --debug --enable-remote main.yaml
cat jobs.tmp-txt >> all-jobs.tmp-txt
USER_ID=`id -u` scalebox app create --debug --enable-remote prep.yaml
cat jobs.tmp-txt >> all-jobs.tmp-txt

# add remote-job
PGHOST=${SERVER_SLAVE} GRPC_SERVER=${SERVER_SLAVE} scalebox app add-remote main.yaml
USER_ID=`id -u` scalebox app add-remote prep.yaml

rm -f *.tmp-txt
