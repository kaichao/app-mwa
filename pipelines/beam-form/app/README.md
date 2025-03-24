# app

```mermaid
flowchart TD
  subgraph beam-form
    A([wait-queue]) --> B([pull-unpack])
    B --> C[beam-make]
    C --> D[down-sample]
    D --> E([fits-redist])
    E --> F[fits-merge]
    F --> G([remote-fits-push])
  end
```

## 一、数据准备

### p419集群

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
  START_MESSAGE=1255803168/p03121_03600 \
  NODES="n-00[01][0-9]|n-002[0-3]" \
  scalebox app create -e p419.env
```

- 数据位于共享存储



## dcu集群

- 数据位于远端共享存储

```sh
  START_MESSAGE=1257617424/p00001_00096/t1257617426_1257617585 \
  TIME_STEP=80 \
  NODES=n-0[123] \
  scalebox app create
```

- 数据位于本地共享存储

