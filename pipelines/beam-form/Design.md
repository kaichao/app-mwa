# beam-form流水线

```mermaid
flowchart TD
  subgraph beam-form
    tar-load --> cube-vtask
    cube-vtask --> pull-unpack
    pull-unpack --> beam-make
    beam-make --> down-sample
    down-sample --> fits-redist
    fits-redist --> fits-merge
    fits-merge --> fits24ch-save
    fits24ch-save --> fits24ch-unload
  end
```

## 一、功能模块表

| 模块名      | 说明                                                          |
| ---------- | ------------------------------------------------------------ |
| tar-load   | tar文件从外部存储加载到HPC存储, 1257010784/p00001_00960/t1257012766_1257012965/ch109     |
| cube-vtask | 实现group_vtasks流控，限制每节点组的最大vtask数量，按顺序释放消息，管理后续所有HOST-BOUND模块 |
| pull-unpack | 1. 将外部存储/HPC存储的数据拉取到计算节点本地并解包。2. I/O节点将HPC存储的数据解包（标准模块+定制脚本） |
| beam-make   | 按通道的波束合成 |
| down-sample | 波束合成结果fits文件做1/4下采样，降低数据量 |
| fits-redist | 下采样后fits文件，按pointing再分发，以便：1.组内节点按指向合并；2.presto-search节点合并 |
| fits-merge  | 按Pointing合并。合并结果：1.存放到HPC存储；2.存放到本地，通过fits24ch-copy传输到HPC存储或计算节点存储 |
| fits24ch-copy  | 按需，将结果数据拷贝到HPC共享存储或presto计算节点存储 |
| fits24ch-unload | 按需，将结果数据从HPC存储拷贝到外部存储 |
| message-router  |                  |


## 二、模块设计

| num | module_name      | image_name        | std_image|cust_code| input_message     | input_path     | output_message    | output_path    |
| --- | ---------------- | ----------------- | ------ | -----      | ----------------- | ----------------- | ----------------- | ----------------- |
| 1 | tar_load | scalebox/file-copy | Yes   | No    | 1257010784/p00001_00960/t1257012766_1257012965 | | ${input_message} | |
| 2 | cube_vtask | scalebox/agent     | Yes   | No    | 1257010784/p00001_00960/t1257012766_1257012965 | | ${input_message} | |
| 3 | pull_unpack | scalebox/ file-copy     | Yes   | Yes   | 1266932744/1266932986_1266933025_ch118.dat.tar.zst <br/> 1266932744/p00001_00960/1266932986_1266933025_ch118.dat.tar.zst | mwa/tar/1266932744/```$```{input_message} <br/> mwa/tar/p00001_00960/1266932744/```$```{input_message} | ${input_message} | mwa/dat/1266932744/p00001_00960/t1266932986_1266933185/ch118 |
| 4 | beam_make | app-mwa/ mwa-vcstools     | No    | Yes   | 1257010784/p00001_00024/t1257012766_1257012965/ch109 | mwa/dat/${input_message}| ${input_message} |mwa/1ch/${input_message}/p00001.fits |
| 5 | down_sample | app-mwa/ down-sampler   | No    | No    | 1257010784/p00001_00024/t1257012766_1257012965/ch109 |mwa/1ch/${input_message} | ${input_message} | mwa/1chy/1257617424/p00001/t1257012766_1257012965/ch109.fits.zst (non-local)<br/> mwa/1chx/1257617424/p00001_00024/t1257617426_1257617505/ch109/p00001.fits.zst|
| 6 | fits_redist | scalebox/ file-copy     | Yes   | Yes   | 1257010784/p00001_00024/t1257010786_1257010965/ch121 |mwa/1chx/${input_message}|${input_message} |mwa/1chz/1257617424/p00001/t1257012766_1257012965/ch109.fits.zst|
| 7 | fits_merge | app-mwa/ mwa-vcstools    | Yes   | No  | 1257010784/p00023/t1257010786_1257010965 |mwa/1chz/${input_message} | ${input_message} |mwa/24ch/${input_message}.zst|
| 8 | fits24ch_copy | scalebox/ file-copy | Ye  | No  | 1257010784/p00023/t1257010786_1257010965.tar.zst | mwa/24ch/${input_message}| ${input_message} | |
| 9 | fits24ch_unload | scalebox/ file-copy | Yes  | No  | 1257010784/p00023/t1257010786_1257010965.tar.zst | mwa/24ch/${input_message}| ${input_message} | |


### 2.1 cube-vtask

流水处理的vtask流控。

- 处理步骤
  1. 若信号量值为0，自动停止
  2. 信号量```group_vtask_size```自动减一

- 镜像名：scalebox/agent
- 输入消息：
  - 消息体：1257010784/p00001_00960/t1257012766_1257012965
  - 消息头：

- 输出消息：1257010784/p00001_00960/t1257012766_1257012965
- task分发排序(定制排序)
  - sort_tag: {pointing}
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
  - sort_tag: {pointing}
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

### 2.5 fits-redist

### 2.6 fits-merge

### 2.7 fits-push

## 三、信号量/共享变量的设计

信号量表
| category      | sema_name                                                  | initial value    |  comment |
| ------------- | ---------------------------------------------------------- | ---------------- | -------- |
| tar-ready     | tar-ready:1257010784/p00001_00960/t1257010786_1257010985   |                  |          |
| dat-ready | dat-ready:1257010784/p00001_00960/t1257010786_1257010985/ch109 | tar.zst打包文件数 |          |
| dat-done   | dat-done:1257010784/p00001_00960/t1257010786_1257010985/ch109 | 指向组处理次数     |          |
| fits-done     | fits-done:1257010784/p00001_00024/t1257010786_1257010985   |       24         |          |
| pointing-done | pointing-done:1257010784/p00001                            |  时间区段长度      |          |
| task_progress | task_progress:beam-make:g01h00                             |                  |          |
| capacity-presto-search | capacity-presto-search:h0000                   |  |计算节点上presto-search的vtask数 |

变量表
| category           | var_name                             | value                           |
| ------------------ | ------------------------------------ | ------------------------------- |
| pointing_data_root | pointing_data_root:1257010784/p00001 | 10.2.3.4 (计算节点) <br/>  /local_root(本地共享存储) <br/>   remote_user@remote_ip:port/remote-root     |


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

### pointing-data-root
每个指向的数据存储位置。分为三种情况：
- 计算节点本地存储：用所在节点的IP地址表示，缺省目录：/dev/shm/scalebox/mydata/mwa
- 计算集群共享存储：用共享目录表示
- 外部的共享存储：通过ssh表示表示（user@ip-addr:port/remote-dir）

在message-router的```from_module='down-sample'```中，生成、使用该变量。
- 若变量表中不存在该指向对应的变量，则通过以下步骤生成：
  - 优先读取优先队列，生成计算节点；
  - 则依据各个可用共享存储的当前带宽、可用容量等，综合选择一个共享存储。

在message-router的```from_module='fits-merge'```中，使用该变量。若用外部共享存储，通过```fits-push```将生成结果推送过去。

在presto搜索模块中，收到消息后，通过该变量获取波束合成结果。

## 四、message-router设计

| from_module            | input_message            | to_module                    | output_message        |
| ---------------------- | ------------------------ | --------------------------- | ---------------------- |
| (default) | 1257010784 <br/> 1257010784/p00001_00960 <br/> 1257010784/p00001_00960/t1257012766_1257012965 <br/> 1257010784/p00001_00960/t1257012766_1257012965/ch109 | wait_queue <br/> pull_unpack | 1257010784/p00001_00960/t1257012766_1257012965 <br/> 1266932744/p00001_00960/1266933866_1266933905_ch112.dat.tar.zst | 
| wait_queue | 1257010784/p00001_00960/t1257012766_1257012965 | pull_unpack | p00001_00960/1266932744/1266932986_1266933025_ch118.dat.tar.zst |
| pull_unpack | 1266932744/1266932986_1266933025_ch118.dat.tar.zst <br/> 1266932744/p00001_00960/1266932986_1266933025_ch118.dat.tar.zst | beam_make | 1257010784/p00001_00960/t1257012766_1257012965/ch109 |
| beam_make | 1257010784/p00001_00960/t1257012766_1257012965/ch109 | down_sample |  ${input_message} |
| down_sample | 1257010784/p00001_00960/t1257012766_1257012965/ch109 | fits_redist <br/> fits_merge | 1257010784/p00023/t1257010786_1257010965/ch121.fits <br/> 1257010784/p00023/t1257010786_1257010965 |
| fits_redist | 1257010784/p00023/t1257010786_1257010965/ch121.fits | fits_merge |  ${input_message} |
| fits_merge | 1257010784/p00023/t1257010786_1257010965 | fits_push |  ${input_message} |
| fits_push | 1257010784/p00023/t1257010786_1257010965 | (NULL) |  |

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

### fits-push

- 输入消息格式
  - 格式1：
  - 格式2：
- 消息头/环境变量：
- 处理步骤
  - 1.
  - 2.  

