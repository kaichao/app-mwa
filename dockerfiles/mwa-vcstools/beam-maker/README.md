## 模块介绍
将单通道时间区间dat文件序列合并转换为单指向、单通道fits文件。

- 观测的时间戳UTT，放在观测数据集的元数据中（定标目录下的元数据定义文件mb_meta.env）

## 环境变量

| 环境变量              | 变量说明                        |
|-------------------|-----------------------------|
| LOCAL_INPUT_ROOT  | 输入文件单通道dat的本地根目录。设为非空，用于本地计算。  |
| LOCAL_OUTPUT_ROOT | 输出fits文件的本地根目录。设为非空，用于本地计算。 |
| LOCAL_CAL_ROOT    | 定标文件的本地根目录。设为非空，用于本地计算。     |
| KEEP_SOURCE_FILE  | 是否保留原始文件。设为no，用于生产运行        |
| KEEP_TARGET_FILE  | 是否保留目标文件。设为no，则用于测试。        |

## 输入消息格式：
  - ${观测号}/${起始时间戳}_${结尾时间戳}/${通道号}/${起始指向号}_${结尾执行好}
  - 例："1257010784/1257010986_1257011185/109/00001_00003"

## 用户应用的退出码
- 0 
## 输出消息格式
- 若退出码为0，则输出与输入消息相同的消息。

## 待完善部分
- 镜像精简，当前基础镜像

## 单观测数据集计算时间预估
- 数据集元数据
  - 24通道
  - 指向数：12960
  - 观测时间：4800秒
- 单通道200秒数据的24指向处理时间为540秒计
  - 总计算时间：24 * (12960/24) * (4800/240) * 540 = 1.4 * 10 ^ 8 秒 = 38880 小时 = 1620 天
  - 单集群24节点，每节点4个DCU卡，则16.875天
  - 若用10个集群240节点，则总计算时间预期可控制在2天以内