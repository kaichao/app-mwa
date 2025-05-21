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
TARGET_URL=cstu0036@60.245.128.14:65010/work2/cstu0036/tmp \
SOURCE_URL=/data2/mydata/mwa/tar DIR_NAME=1255803168 \
scalebox app create
```
### dcu集群

## 二、波束合成计算

## p419集群

- 1440指向
```sh
  START_MESSAGE=1255803168/p05161_06600 \
  PRESTO_APP_ID=102 \
  POINTING_FIRST=yes \
  PRESTO_NODES=a-.+ \
  scalebox app create -e p419.env
```

- 480指向
```sh
  START_MESSAGE=1255803168/p05161_05640 \
  PRESTO_APP_ID=102 \
  NODES=d-.+ \
  TIME_STEP=160 \
  PULL_UNPACK_DIR_LIMIT_GB=/tmp/scalebox/mydata/mwa/dat:90 \
  TARGET_24CH_ROOT=/work1/cstu0036/mydata \
  scalebox app create -e p419.env
```

- 240指向测试
```sh
  START_MESSAGE=1267459328/p00001_00240 \
  PRESTO_APP_ID=102 \
  TIME_STEP=160 \
  PULL_UNPACK_DIR_LIMIT_GB=/tmp/scalebox/mydata/mwa/dat:90 \
  SOURCE_URL=astro@10.100.1.30:10022/data2/mydata \
  TARGET_24CH_ROOT=astro@10.100.1.30:10022/data1/mydata \
  NODES=d-00.+ \
  scalebox app create -e p419.env
```

- 4800指向
```sh
  START_MESSAGE=1255803168/p04921_09320 \
  PRESTO_APP_ID=68 \
  NODES="c-1[23]" \
  scalebox app create -e p419.env
```

- 生产测试，source_url通过p419-soruce.json来指定
```sh
  START_MESSAGE=1255803168/p04681_04920 \
  PRESTO_APP_ID=61 \
  scalebox app create -e p419.env
```


```sh
  START_MESSAGE=1257617424/p00001_00096/t1257617426_1257617585 \
  TIME_STEP=80 \
  NODES="n-000[0-9]|n-001[01]" \
  scalebox app create -e p419.env
```

```sh
  START_MESSAGE=1255803168/p03121_03600/t1255805770_1255807967 \
  NODES="n-00[01][0-9]|n-002[0-3]" \
  scalebox app create -e p419.env
```


## dcu集群

- source_url通过dcu-soruce.json来指定

```sh
  START_MESSAGE=1257617424/p00001_00096 \
  TIME_STEP=80 \
  TIME_END=1257617505 \
  NODES=n-0[023] \
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
