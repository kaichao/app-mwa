#!/bin/bash

current_dir=`dirname $0`

# Extracting relevant information from the input message
# input-message: 1301240224_1301240343_ch117.dat~b00
# sema: dat-ready:1301240224/t1301240225_1301240374/ch118
ch=$(echo "$1" | grep -oP 'ch\K[0-9]+(?=\.dat)')

filename="$current_dir/my-pointings.txt"
# 读取文件的第一行
line=$(head -n 1 "$filename")

# 使用正则表达式提取数字
if [[ $line =~ ([0-9]+)/p([0-9]+)/t([0-9]+)_([0-9]+)\.fits\.zst ]]; then
  ds="${BASH_REMATCH[1]}"
  pointing="${BASH_REMATCH[2]}"
  t0="${BASH_REMATCH[3]}"
  t1="${BASH_REMATCH[4]}"
else
  echo "No match found"
  exit 1
fi
echo "Extracted numbers: $ds, $pointing, $t0, $t1"

sema_name="dat-ready:${ds}/t${t0}_${t1}/ch${ch}"
# Running the `scalebox` command to get a numeric string
n=$(scalebox semaphore countdown "$sema_name")
code=$?

# Checking the exit status of the `scalebox` command
# If there is an error, print an error message and exit with the same code
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore countdown! " >&2 && exit $code 

# Checking if the semaphore is 0
if [ "$n" -eq 0 ]; then
# beam-maker message: 1301240224/1301241875_1301242024/110/00001_00024	
    m="${ds}/${t0}_${t1}/${ch}/${pointing}_${pointing}"
    echo "beam-maker,$m" > $WORK_DIR/messages.txt
fi

echo "finished from-pull-unpack." > $WORK_DIR/custom-out.txt

exit 0
