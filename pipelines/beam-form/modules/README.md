# 模块定义

## 一、读写量计数

放到文件：```${WORK_DIR}/extras.yaml```

### 总读写量
- ```${WORK_DIR}/input_files.txt```、```${WORK_DIR}/output_files.txt```
- ```${WORK_DIR}/task_exec.json```

纪录在t_task_exec的input_bytes、output_bytes中

### 分类读写量

读写量计数：节点名、模块名、字节数

- 共享读/写(依据共享目录)
  - global_input
  - global_output
- tmpfs读/写（/dev/shm）
  - tmpfs_input
  - tmpfs_output
- 本地读/写
  - local_input
  - local_output
- 网络读/写（remote-host,size）
  - network_input
  - network_output
- 总读写
  - input_bytes
  - output_bytes

纪录在t_task_exec的json字段extra ->> iobytes中
