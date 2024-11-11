#!/usr/bin/env bash

i=0
for h in $(cat /tmp/ip_list.txt);do
    dt=$(date +"%H:%M:%S.%N" | cut -b 1-12)
    echo "i=$i Host:$h, [$dt]"
    cmd="ssh -p 50022 $h $*"
    eval "$cmd"
    ((i++))
done

echo
