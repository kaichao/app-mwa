#!/bin/bash

/app/bin/get_host_list.py

# set the variable
scalebox variable set run_cached_pointings ${START_MODE}

redis-cli -h $REDIS_HOST -p $REDIS_PORT DEL $REDIS_QUEUE

# read the host names from host_list.txt and run the command on each host.
# while read -r host; do
#     # create the semaphore
#     echo "host-spare:$host $INIT_SLOTS"
#     scalebox semaphore create host-spare:$host $INIT_SLOTS
#     # if start with no, send message $host:timestamp to redis server $INIT_SLOTS times.
#     # the priority is set from 1 to $INIT_SLOTS. 
#     if [[ $START_MODE = "no" ]]; then
#         for i in $(seq 1 $INIT_SLOTS); do
#             redis-cli -h $REDIS_HOST -p $REDIS_PORT ZADD ${REDIS_QUEUE} $i "$host:$(date +%s)"
#         done
#     fi

# done < host_list.txt

rm ./host_list.txt