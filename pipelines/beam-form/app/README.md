# app

```mermaid
flowchart TD
  subgraph beam-form
    A([wait-queue]) --> B([pull-unpack])
    B --> C[beam-make]
    C --> D[down-sample]
    D --> E([fits-redist])
    E --> F[fits-merge]
    F --> G([fits-push])
  end
```

## 一、数据准备

### p419集群

在scalebox/dockerfiles/files/app-dir-copy目录下

#### 预拷贝文件到共享存储
- 全数据集拷贝

```sh
TARGET_URL=cstu0036@10.100.1.104:65010/work2/cstu0036/tmp \
SOURCE_URL=/data2/mydata/mwa/tar DIR_NAME=1255803168 \
scalebox app create
```

```sh
export SOURCE_URL=/data2/mydata/mwa/tar
export TARGET_URL=cstu0036@10.100.1.104:65010/work2/cstu0036/mydata/mwa/tar

scalebox app run --image-name=hub.cstcloud.cn/scalebox/file-copy:latest 1267459328/1267464090_1267464129_ch127.dat.tar.zst

cd /data2/mydata/mwa/tar
find 1267459328 -type f|scalebox app run --image-name=hub.cstcloud.cn/scalebox/file-copy:latest --slot-regex=h0:2

```


### dcu集群

## 二、波束合成计算

### p419集群全并行

- 1440指向(全并行处理)
```sh
  START_MESSAGE=1255803168/p05161_06600 \
  PRESTO_APP_ID=102 \
  POINTING_FIRST=yes \
  PRESTO_NODES=a-.+ \
  scalebox app create -e p419.env
```


```sh
  START_MESSAGE=1257617424/p00001_00096/t1257617426_1257617585 \
  TIME_STEP=80 \
  NODES="n-000[0-9]|n-001[01]" \
  scalebox app create -e p419.env
```

- 生产测试，source_url通过p419-soruce.json来指定
```sh
  START_MESSAGE=1255803168/p04681_04920 \
  PRESTO_APP_ID=61 \
  scalebox app create -e p419.env
```


### p419集群流式并行

START_MESSAGE=1267459328/p01441_02400/t1267459330_1267464129

以下出错：
START_MESSAGE=1267459328/p02401_03360/t1267459330_1267464129 \

```sh
START_MESSAGE=1267459328/p02401_03360/t1267459330_1267464129 \
  PRESTO_APP_ID=102 \
  NODES=d-0[0].+ \
  PRESTO_NODES= \
  TIME_STEP=200 \
  PULL_UNPACK_LIMIT_GB=120 \
  BEAM_MAKE_FREE_GB='{~n*7+14~}' \
  SOURCE_TAR_ROOT=/work2/cstu0036/mydata \
  TARGET_24CH_ROOT=/work1/cstu0036/mydata \
  scalebox app create -e p419.env
```


- 单次200秒数据处理

```
  TIME_STEP=200 \
  PULL_UNPACK_LIMIT_GB=120 \
  BEAM_MAKE_FREE_GB='{~n*7+14~}' \
```

- 单次160秒数据处理
```
  TIME_STEP=160 \
  PULL_UNPACK_LIMIT_GB=90 \
  BEAM_MAKE_FREE_GB='{~n*5+11~}' \
```

- 单次120秒数据处理
```
  TIME_STEP=120 \
  PULL_UNPACK_LIMIT_GB=65 \
  BEAM_MAKE_FREE_GB='{~n*4+9~}' \
```


### dcu集群

- source_url通过dcu-soruce.json来指定

1267459328/p02401_03360/t1267459330_1267464129

```sh
  START_MESSAGE=1267459328/p00001_00096/t1267459330_1267459409 \
  TIME_STEP=80 \
  NODES=n-0[023] \
  NUM_BEAM_MAKE=2 \
  TARGET_JUMP=root@10.200.1.100 \
  scalebox app create
```


```sh
  START_MESSAGE=1257617424/p00001_00096 \
  TIME_STEP=80 \
  TIME_END=1257617505 \
  NODES=n-0[023] \
  NUM_BEAM_MAKE=3 \
  TARGET_JUMP=root@10.200.1.100 \
  scalebox app create
```


```sh
  START_MESSAGE=1257617424/p00001_00096/t1257617426_1257617585 \
  TIME_STEP=80 \
  NODES=n-0[123] \
  scalebox app create
```

## 新增一个队列原始

```sh
docker exec server_redis_1 redis-cli -h localhost -p 6379 ZADD QUEUE_HOSTS 1.0 10.11.16.79:9876543210

docker exec server_redis_1 redis-cli -h localhost -p 6379 ZADD QUEUE_HOSTS 1.0 10.11.16.79:9876543211
```


## 波束合成结果核对

- 文件按从大到小排列

```sh
find . -type f -exec ls -l {} + | sort -k 5 -nr
```

- 列出一级子目录下的文件数量

```sh
find . -maxdepth 2 -type f | cut -d'/' -f2 | sort | uniq -c | awk '{print $2 ": " $1}'
```


## 重启server端daemon
```sh
LOCAL_ADDR=10.100.1.30 docker compose up -d --no-deps --force-recreate actuator
```