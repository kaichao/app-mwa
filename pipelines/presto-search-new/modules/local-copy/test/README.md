# 测试

## 启动

```sh
app_id=$( scalebox app create | cut -d':' -f2 | tr -d '}' )
```

## 拷贝共享存储文件
```sh
APP_ID=${app_id} scalebox task add --sink-job=local-copy -h source_url=/work1/cstu0036/mydata/mwa/24ch/1255803168-250321/p00100


```