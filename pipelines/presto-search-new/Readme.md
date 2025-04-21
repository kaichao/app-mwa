## 一些想法

    节点分配应该在message-router完成，但数据存放在共享存储时无法实现动态任务分配。不能直接在message-router上设置流控，因此需要一个入口模块作为v_task机制流控限制的对象。

    mr产生任务---->local-wait-queue---->mr分配任务---->local-copy-unpack---->后续

    mr分配任务时遍历所有节点，选择有剩余资源或等待任务最少的并分配。

    两流水线并行时如何解决？

    重要：本地资源有空余时,如何设计机制平衡从不同数据源获取数据？

    我需要在这里写一个设计文档。

# Presto搜索流水线设计

考虑需求，需要支持使用共享存储/本地计算的方案。

## 基于共享存储的流水线

在这种情况下，流水线直接从共享存储上读写数据，并完成计算。

## 基于节点本地计算的流水线

这种情况需要根据数据存放的位置分类处理。

1.数据存放在共享存储上，此时需要先将数据拉取到本地，再完成计算。

2.数据存放在本地节点。此时只需要解压后即可完成后续计算。（或许本地数据无需压缩？）

3.数据存放在远端存储设备。同样需要先将数据传输到本地。

这三种情况可以通过消息的header区分。同时，为了测试流水线的功能，应该能同时使用多种来源的数据进行计算。

对于情况1,3，收到消息后加入local-wait-queue等待。对于情况2，则直接发消息给dedisp-search启动计算。

这三种情况均使用同一模块，依据不同的header决定分支。

在一个指向计算完成时，对于一个存储资源足够多的节点，使用一个随机数决定接下来处理来自共享存储的数据，或是给
redis发消息。

## 功能模块表

- message-router
- shared-wait-queue：所有共享存储上指向的等待队列。
- local-copy: 从共享存储拉取数据到计算节点。
- local-unpack: 解压本地的计算数据。
- rfi-find: 对输入的指向进行解压缩、消干扰。按目前预期，这个模块只用于共享存储上的数据，大规模处理时对同一观测使用预先生成的rfi文件。
- dedisp-search：对指定的DM范围进行消色散与信号搜索
- fold:对指定的DM范围的候选体进行折叠与绘图。
- result-push：将搜索结果推送回头结点。
- remote-push：将结果推送至外部服务器。

## 模块设计

| num | module_name      | image_name        | std_image|cust_code| input_message     | input_path     | output_message    | output_path    |
| --- | ---------------- | ----------------- | ------ | -----      | ----------------- | ----------------- | ----------------- | ----------------- |
| 1 | shared-wait-queue | scalebox/agent     | Yes   | No    | 1257010784/p00001 | | ${input_message} | |
| 2 | fits-merge | app-mwa/ mwa-vcstools    | Yes   | No    | 1257010784/p00023/t1257010786_1257010965 |mwa/1chy/${input_message} | ${input_message} |mwa/24ch/${input_message}.zst|
| 3 | local_copy | scalebox/file-copy | Yes  | Yes   | 1257010784/p00023/ | mwa/24ch/${input_message}| ${input_message} | |
| 4 | rfi-find | app-mwa/ presto-search     | No   |  Yes   | 1257010784/p00001 | mwa/24ch/${input_message} | ${input_message} | mwa/dedisp/${input_message}_rfifind.mask |
| 5 | dedisp-search | app-mwa/ presto-search     | No   | Yes    | 1257010784/p00001/01 | mwa/24ch/1257010784/p00001 | 1257010784/p00001/dm1/group01 | mwa/dedisp/${output_message} |
| 6 | fold | app-mwa/ presto-search     | No   | Yes    | 1257010784/p00001/dm01 | mwa/dedisp/${input_message} | {input_message} | mwa/png/${output_message} |
| 7 | result-push | scalebox/file-copy | Yes  | No    | 1257010784/p00023/ | mwa/24ch/${input_message}| ${input_message} | |
| 8 | remote-push | scalebox/file-copy | Yes  | No    | 1257010784/p00023/ | mwa/24ch/${input_message}| ${input_message} | |

## 信号量设计

| category      | sema_name                                                  | initial value    |  comment |
| ------------- | ---------------------------------------------------------- |  --------------- | -------- |
| pointing-ready | pointing-ready:p00001 | $NUM_FILES | 初值为每个指向的文件数量 |
| pointing-finished | pointing-finished:p00001 | $NUM_GROUPS | 启动dedisp_search的次数 |
| dm-group-ready | dm-group-ready:p00001/dm1 | $calls/$ncall | 取决于消色散配置 |

## 共享变量

| category     | var_name              | value                            |
| -------------------- | ------------------------------ |  -------------- |
| run_cached_pointings | run_cached_pointings:hostname  |  yes/no         |
| local_pointing       | local_pointing:$pointing       |  yes/no         |


## 特殊消息

这里规定一些以Command开头的消息为特殊消息，向message-router发送这些消息用于初始化流水线、将新增节点加入计算、恢复失败的vtask等操作。

| message                     | function                                              |
| --------------------------- | ----------------------------------------------------- |
| Command:init                | 启动消息                                               |
| Command:update-hosts        | 根据节点本地存储大小进行筛选并修改节点名                  |
| Command:add:n-0001          | 为指定节点补充创建信号量与slot，使其能获取信息开始数据处理 |
| Command:stop:n-0001         | 停止向指定节点分发消息，从队列中移除对应信息，准备释放节点 |
| Command:retry:$dataset/$pointing | 重置指定pointing的处理状态并重新运行               |