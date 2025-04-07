# fits-merge


## 一、模块设计
将单指向、24个单通道fits文件合并为单个24通道的fits文件。

### 输入消息格式：
  - ${观测号}/p${指向号}/t${起始时间戳}_${结尾时间戳}
  - 例："1257010784/p00001/t1257010986_1257011185"

### 消息头/环境变量

| 消息头        | 环境变量          | 变量说明                                         |
|------------- | ---------------- | ---------------------------------------------- |
|              | INPUT_ROOT       | 输入文件单通道dat的本地根目录。设为非空，用于本地计算。 |
|              | OUTPUT_ROOT      | 输出fits文件的本地根目录。设为非空，用于本地计算。     |
|              | KEEP_SOURCE_FILE | 是否保留原始文件。设为no，则用于测试。               |

### 用户应用的退出码
- 0 

### 输出消息格式
- 若退出码为0，则输出与输入消息相同的消息。
- 退出码非0，则不输出消息

## 二、模块测试

### 2.1 单task测试
```sh
ret=$(scalebox app create); app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

scalebox task add --app-id=${app_id} --sink-job=fits-merge 1257617424/p00001/t1257617426_1257617505
```


### 2.2 24指向测试
```sh
ret=$(scalebox app create); app_id=$(echo ${ret} | cut -d':' -f2 | tr -d '}')

for p in {00001..00024}; do
scalebox task add --app-id=${app_id} --sink-job=fits-merge 1257617424/p${p}/t1257617426_1257617505
done

```
