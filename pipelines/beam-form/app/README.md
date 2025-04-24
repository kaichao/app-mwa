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

- 数据位于管理服务器

```sh
  START_MESSAGE=1257617424/p00001_00096/t1257617426_1257617585 \
  TIME_STEP=80 \
  NODES="n-000[0-9]|n-001[01]" \
  scalebox app create -e p419.env
```

```sh
  START_MESSAGE=1257617424/p00001_00096/t1257617426_1257617585 \
  TIME_STEP=80 \
  NODES="n-00[01][0-9]|n-002[0-3]" \
  scalebox app create -e p419.env
```

- 生产测试
```sh
  START_MESSAGE=1255803168/p03601_04080 \
  PRESTO_APP_ID=29 \
  scalebox app create -e p419.env
```

```sh
  START_MESSAGE=1255803168/p03121_03600/t1255805770_1255807967 \
  NODES="n-00[01][0-9]|n-002[0-3]" \
  scalebox app create -e p419.env
```
- 多组测试
```sh
  START_MESSAGE=1255803168/p03601_04080 \
  NODES="n-00[0-6][0-9]|n-007[01]" \
  scalebox app create -e p419.env
```

```sh
  START_MESSAGE=1255803168/p03601_04080 \
  TIME_STEP=80 \
  NODES="n-00[0-3][0-9]|n-004[0-7]" \
  scalebox app create -e p419.env
```

- 数据位于共享存储

## dcu集群

- 数据位于远端共享存储

```sh
  START_MESSAGE=1257617424/p00001_00048 \
  TIME_STEP=80 \
  TIME_END=1257617505 \
  NODES=n-0[123] \
  TARGET_JUMP=root@10.200.1.100 \
  scalebox app create
```


```sh
  START_MESSAGE=1257617424/p00001_00096/t1257617426_1257617585 \
  TIME_STEP=80 \
  NODES=n-0[123] \
  scalebox app create
```

```sh
  START_MESSAGE=1257617424/p00001_00048 \
  TIME_STEP=80 \
  NODES=n-0[123] \
  scalebox app create
```

```sh
  START_MESSAGE=1257617424/p00001_00048/t1257618626_1257622223 \
  TIME_STEP=80 \
  NODES=n-0[123] \
  scalebox app create
```

- 数据位于本地共享存储

