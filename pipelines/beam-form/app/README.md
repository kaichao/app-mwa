# app

```mermaid
flowchart TD
  subgraph beam-form
    A([cube-vtask]) --> B([pull-unpack])
    B --> C[beam-make]
    C --> D[down-sample]
    D --> E([fits-redist])
    E --> F[fits-merge]
    F --> G([fits-push])
  end
```

## 一、数据拷贝

### p419集群

在scalebox/dockerfiles/files/app-dir-copy目录下

#### 预拷贝tar文件到HPC存储
- 全数据集拷贝

```sh
TARGET_URL=cstu0036@10.100.1.104:65010/work2/cstu0036/tmp \
SOURCE_URL=/data2/mydata/mwa/tar DIR_NAME=1255803168 \
scalebox app create
```

```sh
export SOURCE_URL=/data2/mydata/mwa/tar
export TARGET_URL=cstu0036@60.245.128.14:65010/work2/cstu0036/mydata/mwa/tar

scalebox app run --image-name=hub.cstcloud.cn/scalebox/file-copy:latest 1267459328/1267464090_1267464129_ch127.dat.tar.zst

cd /data2/mydata/mwa/tar
find 1302106648 -type f|sort|scalebox app run --image-name=hub.cstcloud.cn/scalebox/file-copy:latest --slot-regex=h0:4

```


#### 结果文件拷贝
```sh

export SOURCE_URL=cstu0036@60.245.128.14:65010/work1/cstu0036/mydata/mwa/24ch
export TARGET_URL=/data1/mydata/mwa/24ch
ssh login1 'cd /work1/cstu0036/mydata/mwa/24ch && find 1265983624-250703 -type f' | sort | scalebox app run --image-name=hub.cstcloud.cn/scalebox/file-copy:latest --slot-regex=h0:4

```


### dcu集群



## 二、波束合成计算

### p419集群全并行

- 1440指向(全并行处理)

```sh
START_MESSAGE=1302106648/p01321_01800/t1302106649_1302111446 \
  PRESTO_APP_ID=139 \
  PRELOAD_MODE=none \
  NODES=d-0[01].+ \
  PRESTO_NODES=c.+ \
  TIME_STEP=200 \
  PULL_UNPACK_LIMIT_GB=120 \
  BEAM_MAKE_FREE_GB='{~n*7+14~}' \
  scalebox app create -e p419.env
```

```sh
  START_MESSAGE=1255803168/p05161_06600 \
  PRESTO_APP_ID=102 \
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

```sh
START_MESSAGE=1255803168/p00001_00960 \
  PRESTO_APP_ID=102 \
  POINTING_FILE=pointings-250618.txt \
  NODES=d-0[01].+ \
  PRESTO_NODES= \
  TIME_STEP=160 \
  PULL_UNPACK_LIMIT_GB=90 \
  BEAM_MAKE_FREE_GB='{~n*5+11~}' \
  SOURCE_TAR_ROOT=/work2/cstu0036/mydata \
  TARGET_24CH_ROOT=/work2/cstu0036/mydata \
  scalebox app create -e p419.env
```

START_MESSAGE=1267459328/p07681_09016/t1267459330_1267464129 \

START_MESSAGE=1265983624/p01201_01680/t1265983626_1265988429 \

START_MESSAGE=1302106648/p00001_00480/t1302106649_1302111446 \

```sh
START_MESSAGE=1265983624/p01201_01680/t1265983626_1265988429 \
  PRESTO_APP_ID=102 \
  PRELOAD_MODE=none \
  NODES=d-0[01].+ \
  PRESTO_NODES= \
  TIME_STEP=160 \
  PULL_UNPACK_LIMIT_GB=95 \
  BEAM_MAKE_FREE_GB='{~n*5+11~}' \
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

1302106648/p00001_00096/t1302106649_1302106888

```sh
  START_MESSAGE=1302106648/p00001_00096/t1302106649_1302106888 \
  SOURCE_TAR_ROOT=scalebox@159.226.237.136:10022/raid0/tmp \
  PRELOAD_MODE=none \
  TIME_STEP=80 \
  NODES=n-0[012] \
  NUM_BEAM_MAKE=2 \
  TARGET_JUMP=root@10.200.1.100 \
  scalebox app create
```

- singularity
```sh
  START_MESSAGE=1267459328/p00001_00096/t1267459330_1267459409 \
  TIME_STEP=80 \
  NODES=n-0[012] \
  NUM_BEAM_MAKE=2 \
  TARGET_JUMP=root@10.200.1.100 \
  scalebox app create -e dcu.env
```


```sh
  START_MESSAGE=1257617424/p00001_00096 \
  TIME_STEP=80 \
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
