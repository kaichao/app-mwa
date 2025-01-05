# down-sample

## 一、模块设计

### 输入消息格式：
  - ${观测号}/p${起始指向号}_${结尾指向号}/t${起始时间戳}_${结尾时间戳}/ch${通道号}
  - 例："1257010784/p00001_00024/t1257010986_1257011185/ch109"


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
