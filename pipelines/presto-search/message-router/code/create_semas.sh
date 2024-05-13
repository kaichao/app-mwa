#!/bin/bash

#scalebox@159.226.237.136/raid0/tmp/mwa/24ch-240408~1257010784/p00001	
m=$1

# check if the message is legal (i.e. contains one ~)
if [ `echo $m | grep -c "~"` -ne 1 ]; then
    echo "[ERROR] message format error: $m"
    exit 1
fi
# get the pointing dir
p=`echo $m | awk -F "~" '{print $2}'`

sema1="fits-24ch-presto-ready:$p"
echo "$sema1"
scalebox semaphore create $sema1 32

file_path=/app/bin/MWA_DDplan.txt
if [ -f $file_path ]; then
    total_lines=$(wc -l < $file_path)
    echo $total_lines

    # the first line is the header, skip it
    for ((i=2; i<=$total_lines; i++)); do
        line=$(sed -n "${i}p" $file_path)
        calls=$(echo $line | awk '{print $8}')
        j=$((i - 1))
        echo "line $j: $calls"
        sema="dm-group-ready:$p/dm$j"
        echo "$sema"
        scalebox semaphore create $sema $calls
    done

    echo "total lines: ${MAX_LINENUM}"
    sema2="fits-24ch-dedisp-completed:$p"
    echo "$sema2"
    scalebox semaphore create $sema2 ${MAX_LINENUM}

else
    echo "DDplan file not found: $file_path"
    exit 2
fi
