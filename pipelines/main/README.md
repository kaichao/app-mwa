# 波束合成流水线

## 流水线功能

- 原始数据传输的三种方式
  - 远端集群
    - 先传输到本地集群
    - 通过跳板，直接传输到计算节点
  - 本地集群
    - 传输到计算节点



## 流水线功能性目录
- DIR_DAT：原始dat文件目录，数据规模大，流式传输的计算节点本地存储，并解包，阶段性计算完成就删除
- DIR_CAL：定标文件目录，数据量在100MB以内，可在计算开始前，一次性拷贝到共享存储或本地存储；
- DIR_1CH：单通道fits目录，make_beam输出的单通道的fits目录。临时存储，优先放在本地存储，用完即删
- DIR_24CH：24通道fits目录，splice_psrfits合成的24通道的fits文件目录。波束合成的结果，优先放在本地存储，可能是单通道fits目录下文件相互网络传输后，再生成的结果，在后续单脉冲搜索、周期脉冲完成后才会删除。

- 以上这四个目录的数据访问特性各不相同。

## 功能性目录结构定义

以上的4组功能性目录各自独立。考虑到多数据集处理支持，建议功能性目录之下设数据集目录

### DIR_DAT目录

```
${DIR_DAT}/
      /1252177744/
      /1257010784/
            1257010784_1257011080_ch110.dat
```

### DIR_CAL目录
定标目录，可以将数据集时间戳UTT、指向文件名、观测数据集的起始时间戳、终止时间戳等，都放在数据集目录中
```
${DIR_CAL}
      /1252177744/
      /1257010784/
            /pointings.txt
            /mb_meta.env      以环境变量形式表示的元数据
            /metafits_ppds.fits
```

mb_meta.env文件
```
UTT=2019-11-05T17:43:25.00
```

### DIR_1CH目录
  单通道fits文件目录

```
${DIR_1CH}
      /1252177744/
      /1257010784/
            /1257010986_1257011185/
                  /00001/
                        /ch109.fits
```

### DIR_24CH目录
24通道fits文件目录

```
${DIR_24CH}
      /1252177744/
      /1257010784/
            /1257010986_1257011185/
                  /00001.fits
```

## 单次计算数据量估算

|  dat文件数据长度（秒） | DIR_DAT单通道<br/>数据量（GiB） | DIR_DAT 24通道<br/>数据量（GiB） | DIR_1CH 24指向<br/>数据量(GiB) | DIR_1CH 48指向<br/>数据量(GiB)  | DIR_1CH 72指向<br/>数据量(GiB) | 12000指向<br/>结果数据总量(GiB) |
|  ----  | ---- | ---- | ---- | ---- | ---- | ---- |
| 30  | 9.16   | 219.73   | 2.88  | 5.76  | 8.64  | 1440  |
| 60  | 18.31  | 439.45   | 5.76  | 11.52 | 17.28 | 2880  |
| 90  | 27.47  | 659.18   | 8.64  | 17.28 | 25.92 | 4320  |
| 120 | 36.62  | 878.91   | 11.52 | 23.04 | 34.56 | 5760  |
| 150 | 45.78  | 1098.63  | 14.4  | 28.8  | 43.2  | 7200  |
| 180 | 54.93  | 1318.36  | 17.28 | 34.56 | 51.84 | 8640  |
| 210 | 64.09  | 1538.09  | 20.16 | 40.32 | 60.48 | 10080 |
| 240 | 73.24  | 1757.81  | 23.04 | 46.08 | 69.12 | 11520 |
| 270 | 82.40  | 1977.54  | 25.92 | 51.84 | 77.76 | 12960 |
| 300 | 91.55  | 2197.27  | 28.8  | 57.6  | 86.4  | 14400 |
| 330 | 100.71  | 2416.92  | 31.68  | 63.36  | 95.04  | 15840 |
| 360 | 109.86  | 2636.64  | 34.56  | 69.12  | 103.68 | 17280 |


## message-router的消息排序

message-router消息多，容易被堵塞。若没有有效排序，会导致流式处理不顺畅。通过设置sort_tag，使得消息路由按照一定优先级处理消息。

按以下优先顺序设置：
1. local-tar-pull: 对unpack消息设置最高优先级（'0000'），unpack消息数量较少，尽快从内存cache中将原始数据解包到本地SSD；
2. fits-merger：对fits-24ch-push消息也设置次高优先级（'1111'），fits-24ch-push消息也相对较少，在有限时段中产生，尽快启动将文件传输到存储节点；
3. beam-maker：对down-sampler消息设置再次优先级('2222')，消息数量多，在较长时段都会产生。可尽快将内存cache的占用减少；

## 操作步骤

- 将cal目录拷贝到本机缓存中

```sh

scp -r cal node3:/dev/shm/scalebox/mydata/mwa/

```

- 启动应用
```sh
make
```

## node-agent/file-copy的镜像转为singularity

```sh

mkdir -p ~/singularity/scalebox/

date
singularity build -F ~/singularity/scalebox/file-copy.sif  docker-daemon://hub.cstcloud.cn/scalebox/file-copy:latest
singularity build -F ~/singularity/scalebox/node-agent.sif docker-daemon://hub.cstcloud.cn/scalebox/node-agent:latest
date

ssh login1 mkdir -p singularity/scalebox/
scp  ~/singularity/scalebox/file-copy.sif login1:singularity/scalebox/
scp  ~/singularity/scalebox/node-agent.sif login1:singularity/scalebox/


```

- 模块运行平均时间：

| 模块名        | task数量 | 平均时间 | 累计时间  | slot数量 | slot均时  |
| -------------- | ------ | ------ | -------- | ------ | ------- |
| local-tar-pull | 192    | 85.99  | 16510.94 | 3      | 5503.65 |
| unpack         | 192    | 18.22  | 3498.93  | 3      | 1166.31 |
| beam-maker     | 384    | 187.81 | 72120.58 | 12     | 6010.05 |
| down-sampler   | 9216   | 1.21   | 11130.33 | 3      | 3710.11 |
| fits-redist    | 6144   | 1.27   | 7819.48  | 3      | 2606.49 |
| fits-merger    | 384    | 9.14   | 3510.67  | 3      | 1170.22 |
| fits24ch-push  | 384    | 8.99   | 3453.67  | 3      | 1151.22 |


- 监控的目录容量
  - /dev/shm/scalebox/mydata/mwa/1ch
  - /dev/shm/scalebox/mydata/mwa/1chx
  - /dev/shm/scalebox/mydata/mwa/24ch
  - /dev/shm/scalebox/mydata/mwa/tar
  - /tmp/scalebox/mydata/mwa/dat
- 监控分区自由空间
  - /tmp/scalebox/mydata
  - /dev/shm/scalebox/mydata


## 流水线优化的参数选择
### 数据量分析
- 40s的数据量
  - 波束合成后fits：48.92MB/波束
  - 采样后（1:4）fits：12.23MB/波束
  - 采样后再压缩（压缩因子0.9）fits：11.10MB/波束
- 数据量估算(以40s计)
  - 单节点4个DCU
  - 每轮暂存数据： 4 DCU * 24 波束/DCU * 11.1 MB/波束 =  1066 MiB
  - 计算存储
    - 波束合成：48.92 MB/波束 * 24 波束/DCU * 4 DCU = 4696 MiB
    - 下采样：单实例，对单波束文件（大小不大于50MB倍数）的处理，所需内存缓存在1GB以内，设置阈值为2GB/1GB
    - 再分发：考虑不同节点的不均匀，设置流控3GB/2GB(波束合成、合并所需计算存储 + 1)
    - 合并：单实例，12.23 MB / 波束 * 24波束 = 293.5 MiB

### 主要流控参数
- cluster-dist模块
  - dir_limit_gb：读缓存大小，可设定为2048~5120（流式，2TB~5TB）
- pull-unpack模块
  - dir_free_gb：UNPACK_DIR_FREE_GB。针对40s数据包，解压后约13GB，考虑到临时空间需求，可设定为15~20
  - task_progress_diff：同步各节点上打包文件数，以免因后续处理速度差别，而耗完本地SSD容量。每个文件解包后约12.5GB，该值固定为120，即不超过3个文件；
  - BW_LIMIT: 最大带宽，缺省为25m
- beam-maker模块
  - dir_free_gb: ${BEAM_MAKER_DIR_FREE_GB}，为主要流控参数。用流控表达式表示。
    - 针对单次150秒数据，可取值为{~n*5+8~}，其中单次24指向150秒数据产生的中间结果约4450MB，取值为5；考虑到其他模块中间存储、保留存储的需求，首个容器的取值为8
  - task_progress_diff: 取值范围96~288（1~3组），缺省值可取为144.
- down-sampler模块
  - ？
- fits-redist
  - ?
- fits-merger
  - ?

### 节点数少于24节点，单批次处理指向数
- 单批次中间存储，需存储n-1个通道的待合并数据
- 单批次处理指向数取决于本机/dev/shm中的容量
