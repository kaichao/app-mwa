# beam-maker性能测试

基于共享存储的性能测试

## 一、数据准备

- 到p419
```sh
rsync -av -e "ssh -p 10022" /raid0/tmp/mwa/new-tar1257010784/1257010784/1257010786_1257010815_ch1* kaichao@60.245.128.60:/data/sata/1257010784

for ch in {111..132};do echo $ch; cd ch${ch}/1257010786_1257010815/; zstd -dc ../../1257010786_1257010815_ch${ch}.dat.tar.zst|tar xf -;cd -;done

```

- 到dcu

```sh
rsync -av /raid0/tmp/mwa/new-tar1257010784/1257010784/1257010786_1257010815_ch1* root@223.193.33.31:/raid0/scalebox/mydata/mwa/dat

for ch in {109..132}; do echo $ch; mkdir -p ch${ch}/1257010786_1257010815/; cd ch${ch}/1257010786_1257010815/; zstd -dc ../../1257010786_1257010815_ch${ch}.dat.tar.zst|tar xf -;cd -; done

```


解压后文件结构
- ${CLUSTER_DATA_ROOT}/mwa/dat/1257010784/ch119/1257010786_1257010815/1257010784_1257010847_ch119.dat

## 二、实验设计
### 算法模块
- beam-maker

- fits-merger

- message-router

### 实验参数
- 1~12960指向（24*540）的波束合成
- 2、3、4、6、8、12、24节点，每节点4个DCU

### 数据量估算

- dat文件：152.61GiB/219.72GiB
- 24指向、24个的30秒单通道文件：
- 24指向、30秒的24通道文件：
- 12960指向、30秒、24通道文件：

### 运行时间估计
24指向、30秒，本地存储：运行时间约：80秒
12960指向（540任务），本地存储：运行时间累计：43200秒

24节点（96 DCU）：450秒


## 三、实验流水线

- 本地存储计算（基于main）

- 共享存储计算

## 四、实验结果

节点数、beam-maker共享、down-sampler共享、fits-merger共享、beam-maker本地、fits-merger本地、down-sampler本地、fits-redist本地

节点数：2、3、4、6、8、12、24

模块总运行时间、单任务平均时间

