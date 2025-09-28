# beam-make

## 一、模块设计
将单通道、给定时间区间的dat文件序列合并转换为单指向、单通道fits文件。

- 观测的时间戳UTT，通过python脚本gps2utc.py获取

### 输入消息格式：
  - ${观测号}/p${起始指向号}_${结尾指向号}/t${起始时间戳}_${结尾时间戳}/ch${通道号}
  - 例："1257010784/p00001_00024/t1257010986_1257011185/ch109"

### 消息头/环境变量

| 消息头             | 环境变量              | 变量说明                   |
|------------------ | ------------------ | --------------------------- |
|                   | INPUT_ROOT  | 输入文件单通道dat的本地根目录。设为非空，用于本地计算。|
|                   | OUTPUT_ROOT | 输出fits文件的本地根目录。设为非空，用于本地计算。    |
|                   | CAL_ROOT    | 定标文件的本地根目录。设为非空，用于本地计算。        |
|                   | POINTING_FILE     | 指向文件的名称，缺省为pointings.txt。       |
|                   | KEEP_SOURCE_FILE  | 是否保留原始文件。设为no，则用于测试。        |
|                   | KEEP_TARGET_FILE  | 是否保留目标文件。设为no，则用于测试。        |
| pointing_range    | POINTING_RANGE    | 指向范围，用于确定输入数据的目录。            |


### 用户应用的退出码
- 0 
### 输出消息格式
- 若退出码为0，则输出与输入消息相同的消息。

## 二、功能测试

```sh
ret=$(scalebox app create); app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

scalebox task add --app-id=${app_id} --sink-module=beam-make -h pointing_range=p00001_00960 1257617424/p00001_00024/t1257617426_1257617505/ch109

```

## 三、性能测试

### 3.1 单个task测试
```sh
ret=$(scalebox app create); app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

scalebox task add --app-id=${app_id} --sink-module=beam-make -h pointing_range=p00001_00960 1257617424/p00001_00024/t1257617426_1257617465/ch109

```

### 3.2 全波束24task测试
```sh
ret=$(scalebox app create combined.yaml); app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

for ch in {109..132}; do
scalebox task add --app-id=${app_id} --sink-module=beam-make -h pointing_range=p00001_00048 1257617424/p00001_00024/t1257617426_1257617505/ch${ch}
done
```

## 单观测数据集计算时间预估
- 数据集元数据
  - 24通道
  - 指向数：12960
  - 观测时间：4800秒
- 单通道200秒数据的24指向处理时间为540秒计
  - 总计算时间：24 * (12960/24) * (4800/240) * 540 = 1.4 * 10 ^ 8 秒 = 38880 小时 = 1620 天
  - 单集群24节点，每节点4个DCU卡，则16.875天
  - 若用10个集群240节点，则总计算时间预期可控制在2天以内


## beam-maker中闰秒文件下载

```sh
https://hpiers.obspm.fr/iers/bul/bulc/Leap_Second.dat
```