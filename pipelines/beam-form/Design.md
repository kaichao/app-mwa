# beam-form流水线


## 一、功能模块表

- message-router
- wait-queue：实现group_vtasks的流控模块，每节点组的最大并行数作为流控机制，按顺序释放消息，管理后续所有HOST-BOUND模块。
- pull-unpack：按通道，将远端存储/本地共享存储将数据拉取到计算节点的本地存储，并解包。（标准模块+定制脚本）
- beam-make：按通道的波束合成。
- down-sample：波束合成结果fits文件做1/4下采样，降低数据量
- fits-redist：下采样后fits文件，按pointing再分发，以便按指向合并；
- fits-merge：按Pointing归并，合并结果：1.存放到共享存储；2.存放到本地（供presto-search流水线拉取；通知presto-search流水线来拉取）
- remote-fits-push：按需，将结果数据拷贝到外部存储

## 二、模块设计

| num | module_name      | image_name        | std_image|cust_code| input_message     | input_path     | output_message    | output_path    |
| --- | ---------------- | ----------------- | ------ | -----      | ----------------- | ----------------- | ----------------- | ----------------- |
| 1 | wait_queue | scalebox/agent     | Yes   | No    | 1257010784/p00001_00960/t1257012766_1257012965 | | ${input_message} | |
| 2 | pull_unpack | scalebox/ file-copy     | Yes   | Yes   | 1266932744/1266932986_1266933025_ch118.dat.tar.zst <br/> 1266932744/p00001_00960/1266932986_1266933025_ch118.dat.tar.zst | mwa/tar/1266932744/```$```{input_message} <br/> mwa/tar/p00001_00960/1266932744/```$```{input_message} | ${input_message} | mwa/dat/1266932744/p00001_00960/t1266932986_1266933185/ch118 |
| 3 | beam_make | app-mwa/ mwa-vcstools     | No    | Yes   | 1257010784/p00001_00024/t1257012766_1257012965/ch109 | mwa/dat/${input_message}| ${input_message} |mwa/1ch/${input_message}/p00001.fits |
| 4 | down_sample | app-mwa/ down-sampler   | No    | No    | 1257010784/p00001_00024/t1257012766_1257012965/ch109 |mwa/1ch/${input_message} | ${input_message} | mwa/1chy/1257617424/p00001/t1257012766_1257012965/ch109.fits.zst (non-local)<br/> mwa/1chx/1257617424/p00001_00024/t1257617426_1257617505/ch109/p00001.fits.zst|
| 5 | fits_redist | scalebox/ file-copy     | Yes   | Yes   | 1257010784/p00001_00024/t1257010786_1257010965/ch121 |mwa/1chx/${input_message}|${input_message} |mwa/1chz/1257617424/p00001/t1257012766_1257012965/ch109.fits.zst|
| 6 | fits_merge | app-mwa/ mwa-vcstools    | Yes   | No    | 1257010784/p00023/t1257010786_1257010965 |mwa/1chz/${input_message} | ${input_message} |mwa/24ch/${input_message}.zst|
| 7 | remote_fits_push | scalebox/ file-copy | Ye   | No    | 1257010784/p00023/t1257010786_1257010965.tar.zst | mwa/24ch/${input_message}| ${input_message} | |


### 2.1 wait-queue

流水处理的vtask流控。

- 处理步骤
  1. 若信号量值为0，自动停止
  2. 信号量```group-running-vtask```自动减一

- 镜像名：scalebox/agent
- 输入消息：
  - 消息体：1257010784/p00001_00960/t1257012766_1257012965
  - 消息头：

- 输出消息：1257010784/p00001_00960/t1257012766_1257012965
- task分发排序(定制排序)
  - sorted_tag: {pointing}
  - group_regex: ^([0-9]+)/p([0-9]+)_[0-9]+/t([0-9]+)
  - group_index: 1,2,3
- 流控参数：
  - group_running_vtasks: 3

- 主要环境变量

- task-timeout

### 2.2 pull-unpack
- 输出消息体：1266932744/p00001_00960/1266932986_1266933025_ch118.dat.tar.zst

- 输入消息头：
    - target_subdir:1266932746_1266932945

- 输入目录：
- 输出目录：

- task分发排序(定制排序)
  - sorted_tag: {pointing}
  - group_regex: ^([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]+)
  - group_index: 1,2,3
- 流控参数：
  - dir_limit_gb: mwa/tar~3000
  - 同步流控：


- 主要环境变量
  - 
- task-timeout

### 2.3 beam-make

### 2.4 down-sample

- 主要环境变量
  - LOCAL_COMPUTE: 本地计算模式，取值'yes'。若为非本地计算模式，需对最终输出的fits文件的目录结构做调整。

### 2.5 fits-redist

### 2.6 fits-merge

### 2.7 remote-fits-push

## 三、信号量/共享变量的设计

信号量表
| category      | sema_name                                                  | initial value    |  comment |
| ------------- | ---------------------------------------------------------- |  --------------- | -------- |
| tar-ready     | tar-ready:1257010784/p00001_00960/t1257010786_1257010985   |                  |          |
| dat-ready | dat-ready:1257010784/p00001_00960/t1257010786_1257010985/ch109 | tar.zst打包文件数 |          |
| dat-done   | dat-done:1257010784/p00001_00960/t1257010786_1257010985/ch109 | 指向组处理次数     |          |
| fits-done     | fits-done:1257010784/p00001_00024/t1257010786_1257010985   |       24         |          |
| pointing-done | pointing-done:1257010784/p00001                            |  时间区段长度      |          |
| progress-beam-make | progress-beam-make:g01h00                             |                  |          |
| capacity-presto-search | capacity-presto-search:h0000                   |  |计算节点上presto-search的vtask数 |

变量表
| category     | var_name        | value                            |
| ------------ | --------------- |  ------------------------------- |
| pointing     | pointing:00001  |  g01h00                          |


### tar-ready

标识外部存储到共享存储的归档文件已全部就绪。从外部存储流式将数据传输到集群共享存储。

若单次处理200秒数据，24个channel，tar文件打包40秒数据。则单次处理归档归档文件数量为：24*200/40=120

- 信号量初值：120
- 信号量操作：完成一个tar的copy，信号量减一
- 信号量触发：值为0，给wait-queue发消息

### dat-ready

标识beam-make涉及的单通道dat文件已就绪。

若beam-make单次处理200秒数据，则涉及有5个tar文件

- 信号量初值：典型值为5（200秒）
- 信号量操作：每个打包文件解包完成，对应信号量减一
- 信号量触发：按指向给beam-maker发送消息

### dat-done

用于流式处理过程中，原始科学数据删除的标志位。

标识beam-make单组dat文件需执行的波束合成次数，作为删除本地dat文件的条件。

若指向为p00001_00960，单批次24指向，则初值为960/24=40

- 信号量初值：40
- 信号量操作：每次beam-make处理完成，对应信号量减一
- 信号量触发：按指向给beam-maker发送消息

### fits-done

- 信号量初值：24，24通道
- 信号量操作：每个单通道的fits完成后
- 信号量触发：按指向给beam-make发送消息

### pointing-done
若单次处理200秒数据，则4800秒数据需处理24次。

- 信号量初值：24
- 信号量操作：
- 信号量触发：


## 四、message-router设计

| from_module            | input_message            | to_module                    | output_message        |
| ---------------------- | ------------------------ | --------------------------- | ---------------------- |
| (default) | 1257010784 <br/> 1257010784/p00001_00960 <br/> 1257010784/p00001_00960/t1257012766_1257012965 | wait_queue <br/> remote_tar_pull | 1257010784/p00001_00960/t1257012766_1257012965 <br/> 1266932744/p00001_00960/1266933866_1266933905_ch112.dat.tar.zst | 
| wait_queue | 1257010784/p00001_00960/t1257012766_1257012965 | pull_unpack | p00001_00960/1266932744/1266932986_1266933025_ch118.dat.tar.zst |
| pull_unpack | 1266932744/1266932986_1266933025_ch118.dat.tar.zst <br/> 1266932744/p00001_00960/1266932986_1266933025_ch118.dat.tar.zst | beam_make | 1257010784/p00001_00960/t1257012766_1257012965/ch109 |
| beam_make | 1257010784/p00001_00960/t1257012766_1257012965/ch109 | down_sample |  ${input_message} |
| down_sample | 1257010784/p00001_00960/t1257012766_1257012965/ch109 | fits_redist <br/> fits_merge | 1257010784/p00023/t1257010786_1257010965/ch121.fits <br/> 1257010784/p00023/t1257010786_1257010965 |
| fits_redist | 1257010784/p00023/t1257010786_1257010965/ch121.fits | fits_merge |  ${input_message} |
| fits_merge | 1257010784/p00023/t1257010786_1257010965 | remote_fits_push |  ${input_message} |
| remote_fits_push | 1257010784/p00023/t1257010786_1257010965 | (NULL) |  |

### init()

- 处理步骤
  - 创建```group-running-vtask```的信号量（？）


### (DEFAULT)

接收初始消息，启动数据处理；

- 处理步骤
  1. 若使用了外部存储，则给remote-tar-copy发一组消息，创建信号量```tar-ready```；
  2. 否则，直接给wait-queue发一条消息；针对消息格式1/2，按需创建信号量```pointing-done```，用于与presto-search流水线的同步通信。

- 输入消息
  - 格式1：1257010784。（全数据集处理）
  - 格式2：1257010784/p00001_00960。（指定指向）
  - 格式3：1257010784/p00001_00960/t1257012766_1257012965。

- 输出消息：1257010784/p00001_00960/t1257012766_1257012965

- 消息头/环境变量：
  - BEAM_FORM_ONLY：不创建信号量```pointing-done```

### wait-queue

- 处理步骤
  - 创建后续处理相关信号量（dat-ready/dat-done/fits-ready）
  - 给pull-unpack发一组消息

- 输入消息：1257010784/p00001_00960/t1257012766_1257012965
- 输出消息：
  - 经过remote-tar-copy
    - 消息头： from_dir:p00001_00960/mwa/tar
  - 未经过cluster-tar-copy
    - 消息头： from_dir:mwa/tar

### pull-unpack

- 处理步骤
  - 1. 信号量```dat-ready```减一
  - 2. 若信号量值为0，给```beam-make```发消息

- 输入消息格式：
  - 格式1：
  - 格式2：
- 消息头/环境变量：

### beam-make

- 输入消息
  - 格式1：
  - 格式2：
- 消息头/环境变量：
- 处理步骤
  - 1.
  - 2. 

### down-sample

- 输入消息格式
  - 格式1：
  - 格式2：
- 消息头/环境变量：
- 处理步骤
  - 1.
  - 2. 

### fits-redist

- 输入消息格式
  - 格式1：
  - 格式2：
- 消息头/环境变量：
- 处理步骤
  - 1.
  - 2. 

### fits-merge

- 输入消息格式：1257010784/p00002/t1257010786_1257010985
- 消息头/环境变量：
- 
- 处理步骤
  - 1. 
  - 2. 

### remote-fits-push

- 输入消息格式
  - 格式1：
  - 格式2：
- 消息头/环境变量：
- 处理步骤
  - 1.
  - 2.  

