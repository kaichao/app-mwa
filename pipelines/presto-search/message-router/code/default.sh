#!/bin/bash

#scalebox@159.226.237.136/raid0/tmp/mwa/24ch-240408~1257010784	
m=$1

# check if the message is legal (i.e. contains one ~)
if [ `echo $m | grep -c "~"` -ne 1 ]; then
    echo "[ERROR] message format error: $m"
    exit 1
fi

PB=551
PE=730

echo "ips: " >> ${WORK_DIR}/custom-out.txt
cat /app/bin/ip_list.txt >> ${WORK_DIR}/custom-out.txt

dataset=`echo $m | awk -F "~" '{print $2}'`
echo "total lines: ${MAX_LINENUM}"
file_path=/app/bin/MWA_DDplan.txt
if [ -f $file_path ]; then
    total_lines=$(wc -l < $file_path)
    echo $total_lines

    for ((p = PB; p <= PE; p += 1)); do
        pi=$(printf "%05d" "$p")
        sema="fits-24ch-presto-ready:$dataset/p$pi"
        echo "$sema"
        scalebox semaphore create $sema 32

    # the first line is the header, skip it
        for ((i=2; i<=$total_lines; i++)); do
            line=$(sed -n "${i}p" $file_path)
            calls=$(echo $line | awk '{print $8}')
            j=$((i - 1))
            echo "line $j: $calls"
            sema="dm-group-ready:$dataset/p$pi/dm$j"
            echo "$sema"
            scalebox semaphore create $sema $calls
        done
    
        sema2="fits-24ch-dedisp-completed:$dataset/p$pi"
        echo "$sema2"
        scalebox semaphore create $sema2 ${MAX_LINENUM}
    done
else
    echo "DDplan file not found: $file_path"
    exit 2
fi

# send message to sink job
echo "dir-list,$m" >> ${WORK_DIR}/messages.txt