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



## 操作步骤

- 导入dataset

```sh
CLUSTER=dcu scalebox app create dataset.yaml
```

- 将cal目录拷贝到本机缓存中

```sh

scp -r cal node3:/dev/shm/scalebox/mydata/mwa/

```

- 启动应用
```sh
make
```

## agent/rsync-copy的镜像转为singularity

```sh

mkdir -p ~/singularity/scalebox/
rm -f ~/singularity/scalebox/rsync-copy.sif ~/singularity/scalebox/agent.sif

date
singularity build ~/singularity/scalebox/agent.sif docker-daemon://hub.cstcloud.cn/scalebox/agent:latest
singularity build ~/singularity/scalebox/rsync-copy.sif docker-daemon://hub.cstcloud.cn/scalebox/rsync-copy:latest
date

mkdir -p /raid0/root/singularity/scalebox/
mv -f ~/singularity/scalebox/agent.sif /raid0/root/singularity/scalebox/
mv -f ~/singularity/scalebox/rsync-copy.sif /raid0/root/singularity/scalebox/

```