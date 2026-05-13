# 数据准备

## 一、打包数据拷贝

### p419 -> dcu

- 创建应用
```sh

export SOURCE_URL=scalebox@60.245.128.60:10022/data1/mydata/mwa/tar/1253991112
export SOURCE_JUMP=scalebox@159.226.237.136:22

export TARGET_URL=/raid0/scalebox/mydata/mwa/tar/1253991112

app_id=$(scalebox run --image-name=hub.cstcloud.cn/scalebox/file-copy:latest | cut -d':' -f2 | tr -d '}')
export APP_ID=$app_id

```

- 添加测试任务
```sh
echo 1253991114_1253991153_ch109.dat.tar.zst | scalebox task add
```

- 添加批量任务
```sh
ssh -p 10022 -J scalebox@159.226.237.136:22 scalebox@60.245.128.60 'ls /data1/mydata/mwa/tar/1253991112'| head -n 96 | scalebox task add
```

## 二、打包文件的解包（源路径：文件系统）



```sh
cd pipelines/beam-form/modules/pull-unpack

export SOURCE_URL=/raid0/scalebox/mydata/mwa
export TARGET_URL=/tmp/mydata/mwa

echo 1253991112/1253991114_1253991153_ch109.dat.tar.zst | scalebox run --image-name=hub.cstcloud.cn/scalebox/file-copy:latest 

```

## 三、打包文件的解包（源路径：SSH）

```sh
cd pipelines/beam-form/modules/pull-unpack

export SOURCE_URL=scalebox@60.245.128.60:10022/data1/mydata/mwa
export SOURCE_JUMP=scalebox@159.226.237.136:22
export TARGET_URL=/dev/shm/mydata/mwa

echo 1253991112/1253991114_1253991153_ch109.dat.tar.zst | scalebox run --image-name=hub.cstcloud.cn/scalebox/file-copy:latest 

```

## 四、打包数据集解包

