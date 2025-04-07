## 一些想法

    节点分配应该在message-router完成，但数据存放在共享存储时无法实现动态任务分配。不能直接在message-router上设置流控，因此需要一个入口模块作为v_task机制流控限制的对象。

    mr产生任务---->vhead---->mr分配任务---->local-copy-unpack/（视情况解压到共享存储或本地）---->后续

    入口模块可以承担一些任务？

    同时local-copy-unpack也需要作为后续单节点上的v_task入口完成流控。

    mr分配任务时遍历所有节点，选择有剩余资源或等待任务最少的并分配。

    两流水线并行时如何解决？

    beam-maker---->down-sampler---->fits-redist;

    fits-merger---->(remote-fits-push)---->(local-copy-unpack)---->rfi-find---->后续

    重要：本地资源有空余时，优先通知波束合成流水线向本地推送数据，而非

    我需要在这里写一个设计文档。

# Presto搜索流水线设计

考虑需求，需要支持增加fits-merger与不加fits-merger两种情况，并支持使用共享存储/本地计算的方案。

## 基于共享存储的流水线

在这种情况下，不需要支持fits-merger：在前半流水线完成此功能是可以接受的。



## 基于节点本地计算的流水线

这种情况需要支持fits-merger。

具体而言，有两种情况：其一是与波束合成流水线分别运行的情况，此时数据存放在共享存储上，且已经过波束合成。
此时运行流程为local-copy-unpack-->dedisp-search-->后续，local-wait-queue在local-copy-unpack前，输入消息为
$dataset/$PB_$PE格式。

需要与波束合成流水线共同运行时，设置变量run_cached_pointings为no，从波束合成流水线接收消息后启动对应节点上的
fits-merge，之后顺序执行dedisp-search等模块。接到存放在共享存储的指向时将其加入local-wait-queue等待、

波束合成完成后发送消息，将run_cached_pointings改为yes，开始执行local-wait-queue中的指向。


## 功能模块表

- message-router
- shared-wait-queue：所有共享存储上指向的等待队列。
- fits-merge: 将输入的pointing对应的单通道fits合并为24通道fits文件，合并结果存放到本地。
- local-copy: 从共享存储拉取数据到计算节点。
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
| host-spare | host-spare:10.11.16.80 | $INIT_SLOTS | 取决于本地存储大小 |

## 共享变量

| category     | var_name              | value                            |
| ------------ | --------------------- |  ------------------------------- |
|              | run_cached_pointings  |  yes/no                          |
