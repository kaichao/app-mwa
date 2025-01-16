# beam-make性能测试

## 一、实验环境


## 二、本地存储实验

在实验环境节点上进行，/dev/shm调整为80GB。

### 2.1 实验设计

### 2.2 数据准备
- 调用pull-unpack，在计算节点的本地的/tmp下，分别产生单通道40/80/.../480秒数据，数据量分别约为12.5/25/.../150GB。

在pull-unpack/test目录下，运行以下脚本，准备好不同时段的数据。

- 生成打包文件列表

```sh
cat > tar-files.txt <<EOF
1257617424/1257617426_1257617465_ch109.dat.tar.zst
1257617424/1257617466_1257617505_ch109.dat.tar.zst
1257617424/1257617506_1257617545_ch109.dat.tar.zst
1257617424/1257617546_1257617585_ch109.dat.tar.zst
1257617424/1257617586_1257617625_ch109.dat.tar.zst
1257617424/1257617626_1257617665_ch109.dat.tar.zst
1257617424/1257617666_1257617705_ch109.dat.tar.zst
1257617424/1257617706_1257617745_ch109.dat.tar.zst
1257617424/1257617746_1257617785_ch109.dat.tar.zst
1257617424/1257617786_1257617825_ch109.dat.tar.zst
1257617424/1257617826_1257617865_ch109.dat.tar.zst
1257617424/1257617866_1257617905_ch109.dat.tar.zst
EOF
```

- 运行scalebox app，完成数据准备。n为40秒的文件数(n=1..12)。

```sh
n=9
end=$((1257617426 + n * 40 - 1))

ret=$(SOURCE_URL=scalebox@159.226.237.136:10022/raid0/tmp/mwa/tar1257617424 \
    TARGET_URL=/raid0/scalebox/mydata/mwa/dat-${n} \
    TARGET_SUBDIR=1257617424/p00001_00960/t1257617426_${end}/ch109 \
    scalebox app create)
app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

for f in $(head -n $n "tar-files.txt"); do
    scalebox task add --app-id=${app_id} --sink-job=pull-unpack $f
done


```

### 2.3 实验一：基于本地存储的波束合成

- 波束合成单次处理24个指向；

- 应用参数设置：
  - 本地SSD/内存环境，通过环境变量 ```LOCAL_INPUT_ROOT``` 来设定
    - 本地SSD：```/tmp/scalebox/mydata```
    - 本地内存：```/dev/shm/scalebox/mydata```
  - DCU数量通过环境变量 ```HOSTS``` 来设定，可设定用1/2/3/4个GPU
  - 不同数据量：从40秒到480秒，分别对应dat目录从1到12。

```sh
app_id=$( NUM_SLOTS=2 scalebox app create | cut -d':' -f2 | tr -d '}' )
```

- 添加单个消息
```sh
scalebox task add --app-id=${app_id} --sink-job=beam-make -h pointing_range=p00001_00960 1257617424/p00001_00024/t1257617426_1257617625/ch109
```

- 添加n个消息

每个DCU用8个消息测试，n取值为8/16/24/32

```sh
n=16
for ((i=0; i<n; i++)); do
  start=$((i * 24 + 1))
  end=$((start + 23))
  s=$(printf "%05d_%05d" $start $end)
  scalebox task add --app-id=${app_id} --sink-job=beam-make -h pointing_range=p00001_00960 1257617424/p${s}/t1257617426_1257617625/ch109
done
```

### 2.4 实验二：基于本地存储的波束合成+下采样

```sh
app_id=$( NUM_SLOTS=1 scalebox app create combined.yaml | cut -d':' -f2 | tr -d '}' )
```

```sh
app_id=$( ENABLE_LOCAL_COMPUTE=no NUM_SLOTS=1 scalebox app create combined.yaml | cut -d':' -f2 | tr -d '}' )
```


### 2.5 波束合成的测试结果


## 三、共享存储实验
在生产环境上运行
### 3.1 实验设计
- 单节点实验
- 24

### 3.2 数据准备

### 3.3 基于共享存储的波束合成

### 3.4 基于本地存储的波束合成

- 输入数据：SSD
- 输出数据：tmpfs


## 带宽监控


```sh
#!/bin/bash

# 设置监控的磁盘设备和时间间隔
device="sda"
interval=0.1  # 设定的间隔时间（秒）
sector_size=512  # 扇区大小
correction=0.01135  # 偏差修正，减去的时间（秒）

# 初始化变量
prev_read_sectors=0

while true; do
    start_time=$(date +%s.%N)  # 记录开始时间

    # 获取当前时间的读扇区数
    current_read_sectors=$(awk -v dev="$device" '$3 == dev {print $6}' /proc/diskstats)

    # 如果不是第一次采集，计算变化量和带宽
    if [ -n "$prev_read_sectors" ]; then
        delta_sectors=$((current_read_sectors - prev_read_sectors))
        delta_bytes=$((delta_sectors * sector_size))
        delta_kB=$((delta_bytes / 1024))  # 转换为 KB
        bandwidth=$(echo "scale=2; $delta_kB / $interval" | bc)  # 计算带宽，单位：KB/s

        # 输出结果
        echo "$(date +"%Y-%m-%dT%H:%M:%S.%6N") $delta_sectors, $bandwidth kB/s" >> /dev/shm/diskstats.txt
    fi

    # 更新前一次的读扇区数
    prev_read_sectors=$current_read_sectors

    end_time=$(date +%s.%N)  # 记录结束时间
    elapsed_time=$(echo "$end_time - $start_time" | bc)  # 计算命令执行的时间

    # 计算剩余的时间间隔，确保总间隔为 interval
    sleep_time=$(echo "$interval - $elapsed_time" | bc)

    # 在计算后的 sleep_time 中减去修正的时间偏差（0.01秒）
    corrected_sleep_time=$(echo "$sleep_time - $correction" | bc)

    # 如果修正后的 sleep_time 大于 0，则执行 sleep
    if (( $(echo "$corrected_sleep_time > 0" | bc -l) )); then
        sleep $corrected_sleep_time
    fi
done

```


## 磁盘读写带宽测试

- 单客户端性能测试
- 数据量：20480 MiB
- 带宽单位：MiB/s

| num |  fs     | read-bw|write-bw|   spec                  |
| --- | ------- | ------ | ------ | ----------------------- |
|  1  |  ext4   |  502.0 |  178.7 | SATA SSD                |
|  2  |  tmpfs  | 4382.2 | 2027.2 | DDR4 2666MHz x8         |
|  3  |  ext4   | 8167.5 | 2386.6 | nvme SSD, raid1 x2      |
|  4  |  xfs    | 9946.1 | 3691.3 | nvme SSD, raid0 x2      |
|  5  |  xfs    | 1513.7 | 1924.8 | disk, raid5 x8          |
|  6  |  tmpfs  | 4936.0 | 1500.1 | DDR4 3200MHz x8         |
|  7  | ParaStor|  334.1 |  416.3 | 500+ disks              |

### 测试脚本
- tmpfs
```sh
dd if=/dev/zero of=/dev/shm/testfile bs=1G count=20 
dd if=/dev/shm/testfile of=/dev/null bs=1G
rm -f /dev/shm/testfile
```

- 非tmpfs
  - dir={/tmp,/opt/tmp,/work2/cstu0036}
```sh
dir=/tmp
dd if=/dev/zero of=${dir}/testfile bs=1G count=20 oflag=direct
dd if=${dir}/testfile of=/dev/null bs=1G iflag=direct
rm -f ${dir}
```

### 测试结果

## 四、实验结论


