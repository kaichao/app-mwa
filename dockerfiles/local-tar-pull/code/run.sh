#!/bin/bash

# <user>@<ip-addr>/raid0/tmp/mwa/tar1257010784~1257010784/1257010786_1257010815_ch109.dat.tar.zst~/dev/shm/scalebox/mydata/mwa/tar~01
m0=$1

# remove last characters ~b01
m="${m0%~*}"
batch="${m0##*~}"

/app/share/bin/run.sh $m

# 消息加上批次号，以免在多批次时有同名冲突
file="${WORK_DIR}/messages.txt"
if [[ -f "$file" && $(wc -l < "$file") -eq 1 ]]; then
    line=$(head -n 1 $file)
    echo "${line}~${batch}" > "$file"
fi

exit $?
