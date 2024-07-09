#!/bin/bash

hosts="10.8.1.73 10.8.1.74 10.8.1.75"
i=0
for h in $hosts; do
    dt=$(date +"%H:%M:%S.%N" | cut -b 1-12)
    echo "i=$i Host: $h, [$dt]"
    cmd="ssh $h $*"
    eval "$cmd"
    # echo "all args : $*"
    # ssh $h du -ms /tmp/scalebox
    ((i++))
done
