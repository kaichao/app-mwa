## 模块介绍
将单指向、24个单通道fits文件合并为单个24通道的fits文件。

## 环境变量
  - DIR_1CHX：下采样后的单通道数据目录
  - DIR_24CH：24通道数据目录
## 输入消息格式：
  - ${观测号}/${起始时间戳}_${结尾时间戳}/${指向号}
  - 例："1257010784/1257010986_1257011185/00001"

## 用户应用的退出码
- 0 

## 输出消息格式
- 若退出码为0，则输出与输入消息相同的消息。
- 退出码非0，则不输出消息
