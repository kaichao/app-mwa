# down-sample

## 一、模块设计

### 输入消息格式：
  - ${观测号}/p${起始指向号}_${结尾指向号}/t${起始时间戳}_${结尾时间戳}/ch${通道号}
  - 例："1257010784/p00001_00024/t1257010986_1257011185/ch109"

### 消息头/环境变量

| 消息头      | 环境变量              | 变量说明                                         |
|----------- | -------------------- | ---------------------------------------------- |
|            | LOCAL_INPUT_ROOT     | 输入文件单通道dat的本地根目录。设为非空，用于本地计算。 |
|            | LOCAL_OUTPUT_ROOT    | 输出fits文件的本地根目录。设为非空，用于本地计算。     |
|            | KEEP_SOURCE_FILE     | 是否保留原始文件。设为no，则用于测试。               |
|            | KEEP_TARGET_FILE     | 是否保留目标文件。设为no，则用于测试。               |
|            | ENABLE_LOCAL_COMPUTE | 非本地计算模式，对下采样后目录中的fits文件重新调整目录（本地计算模式，通过fits-redist重新调整目录） |



## 二、模块测试

### 2.1 单task测试
```sh
ret=$(scalebox app create); app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

scalebox task add --app-id=${app_id} --sink-job=down-sample 1257617424/p00001_00024/t1257617426_1257617505/ch109
```

### 2.2 全波束24task测试
```sh
ret=$(scalebox app create); app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

for ch in {109..132}; do
scalebox task add --app-id=${app_id} --sink-job=down-sample 1257617424/p00001_00024/t1257617426_1257617505/ch${ch}
done

```
