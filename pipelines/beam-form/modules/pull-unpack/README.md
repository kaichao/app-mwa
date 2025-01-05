# pull-unpack

## 一、模块设计

### 1.1 输入消息



### 1.2 消息头/环境变量

| 消息头           | 环境变量          | 说明                                 |
| --------------- | ---------------  | ----------------------------------- |
| source_url      | SOURCE_URL       | 本地 或 rsync-over-ssh               |
| target_url      | TARGET_URL       | 本地目录                             |
|                 | JUMP_SERVERS     | ssh的jump_servers                   |
|                 | BW_LIMIT         | 读取数据的最大带宽，10k/10m/10g        |
|                 | KEEP_SOURCE_FILE | 拉取数据后，是否保留原始文件，'yes'/'no' |



### 1.3 输出消息
- 同输入消息

### 1.4 返回错误码


## 二、模块测试

### 2.1 单个文件测试
```sh
SOURCE_URL=/data2/mydata/mwa/tar \
TARGET_URL=/dev/shm/scalebox/mydata/mwa/dat \
START_MESSAGE=1257617424/1257622186_1257622223_ch132.dat.tar.zst \
scalebox app create
```

```sh
ret=$(SOURCE_URL=scalebox@159.226.237.136:10022/raid0/tmp/mwa/tar1266932744 \
TARGET_URL=/tmp/mydata/mwa/dat \
scalebox app create)
app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

scalebox task add --app-id=${app_id} --sink-job=pull-unpack -h target_subdir=1266932744/p00001_00048/t1266937345_1266937543/ch132 1266932744/1266937506_1266937543_ch132.dat.tar.zst

```

### 2.2 文件组测试

- 单通道、单时段
- 引入message-router

- dcu集群
```sh
ret=$(SOURCE_URL=scalebox@159.226.237.136:10022/raid0/tmp/mwa/tar1257617424 \
    TARGET_URL=/raid0/scalebox/mydata/mwa/dat \
    scalebox app create)
app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

for ch in {109..132}; do
scalebox task add --app-id=${app_id} --sink-job=pull-unpack \
    -h target_subdir=1257617424/p00001_00048/t1257617426_1257617505/ch${ch} \
    1257617424/1257617426_1257617465_ch${ch}.dat.tar.zst
scalebox task add --app-id=${app_id} --sink-job=pull-unpack \
    -h target_subdir=1257617424/p00001_00048/t1257617426_1257617505/ch${ch} \
    1257617424/1257617466_1257617505_ch${ch}.dat.tar.zst
done

for ch in {109..132}; do
scalebox task add --app-id=${app_id} --sink-job=pull-unpack \
    -h target_subdir=1257617424/p00001_00048/t1257617506_1257617585/ch${ch} \
    1257617424/1257617506_1257617545_ch${ch}.dat.tar.zst
scalebox task add --app-id=${app_id} --sink-job=pull-unpack \
    -h target_subdir=1257617424/p00001_00048/t1257617506_1257617585/ch${ch} \
    1257617424/1257617546_1257617585_ch${ch}.dat.tar.zst
done

```
### 2.3 全数据集测试


