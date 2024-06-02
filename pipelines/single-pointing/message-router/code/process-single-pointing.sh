#!/bin/bash

source_url="scalebox@159.226.237.136:10022:/raid0/tmp/mwa/tar1301240224"

input=$1
# input="1301240224/p00001/t1301244575_1301244724.fits.zst"

# 使用正则表达式提取数字
if [[ $input =~ ([0-9]+)/p([0-9]+)/t([0-9]+)_([0-9]+)\.fits\.zst ]]; then
  ds="${BASH_REMATCH[1]}"
  pointing="${BASH_REMATCH[2]}"
  num0="${BASH_REMATCH[3]}"
  num1="${BASH_REMATCH[4]}"
else
  echo "No match found"
  exit 1
fi
echo "Extracted numbers: $ds, $pointing, $num0, $num1"

# 1. 生成24个dat-ready的信号量，其值为数据长度
#   例：dat-ready:1301240224/t1301240225_1301240374/ch118
v=$((num1 - num0 + 1))
for ch in {109..132}; do
  sema_name="dat-ready:$ds/t${num0}_${num1}/ch${ch}"
  echo "sema:$sema_name"
  scalebox semaphore create $sema_name $v
done

# 2. 按每个指向结果文件名，生成一个fits-24ch-ready的信号量，其值为24
#   例：fits-24ch-ready:1301240224/p00001/t1301240675_1301240824
sema_name=$(printf "fits-24ch-ready:%s/p%05d/t%s_%s" $ds $pointing $num0 $num1)
echo "sema:$sema_name"
scalebox semaphore create $sema_name 24

# 3. 生成24*n个pull-unpack的消息（gen-pull-unpack-message）
# 输入结果文件名：1301240224/p00001/t1301244575_1301244724.fits.zst
start=$num0
end=$num1
step=30
for ch in {109..132}; do
  # "target_url": "/raid0/scalebox/mydata/mwa/dat/1301240224/ch129/1301241575_1301241724"
  target_url="/raid0/scalebox/mydata/mwa/dat/$ds/ch$ch/${num0}_${num1}"
  echo "target_url:$target_url"
  for ((i=start; i<end; i+=step)); do
    upper=$((i + step - 1))
    if ((upper > end)); then
      upper=$end
    fi
    echo "Interval: $i - $upper"
    # message="1301240224/1301241695_1301241724_ch129.dat.tar.zst~b00"
    message="$ds/${i}_${upper}_ch${ch}.dat.tar.zst~b00"
    echo "msg:$message"
    scalebox task add -h source_url=${source_url} -h target_url=${target_url} --sink-job pull-unpack ${message}
  done
done


# scalebox task add -h source_url=$source_url -h target_url=$target_url $message

