# aggpath 包

## 概述

`aggpath` 包提供了一个用于管理汇聚目录（Aggregated Path）的Go库。它主要用于在分布式系统中管理共享存储路径的分配和释放，将多个共享目录下的存储配额合并，支持更大的存储容量需求，确保多个进程或任务可以安全地共享有限的存储资源。

## 设计原理

### 核心概念

1. **StorageGroup（存储组）**: 一组存储路径的标识，例如 `group0`
2. **Delta（增量）**: 每个存储组对应的存储容量（以GB为单位）
3. **成员路径**: 分配给具体任务或进程的实际路径

### 工作机制

`aggpath` 包使用以下组件协同工作：

1. **信号量（Semaphore）**: 用于跟踪每个存储组的可用存储容量，基于Scalebox的信号量实现
   - 信号量名称格式: `path-free-gb:{storageGroup}`
   - 初始值: 存储组对应的总容量（delta值）

2. **变量（Variable）**: 用于存储已分配的成员路径，基于Scalebox的变量实现
   - 变量名称格式: `member-path:{storageGroup}:{path}`
   - 变量值: 实际分配的成员路径

3. **存储组映射（StorageGroupMap）**: 从配置文件加载的存储组和容量映射

### 工作流程

1. **初始化**: 从配置文件加载存储组和容量映射
2. **分配路径**:
   - 检查请求的路径是否已有分配（通过变量查询）
   - 查找匹配的存储组
   - 检查对应信号量的可用容量
   - 减少信号量值（占用容量）
   - 创建变量记录分配
3. **释放路径**:
   - 查找匹配的存储组
   - 删除分配的目录数据
   - 增加信号量值（释放容量）
   - 删除变量记录

## 安装

```bash
go get beamform/internal/aggpath
```

## 使用方法

### 1. 初始化

#### 容量映射文件的创建及初始化

通过外部命令，创建一个文本文件（如 `my-sema.txt`），表示存储组中的成员目录及可用空间GB数。
```
"path-free-gb:group0:/dir00":500
"path-free-gb:group0:/dir01":400
"path-free-gb:group0:/dir02":300
"path-free-gb:group1:/dir00":300
"path-free-gb:group1:/dir01":200
```

每行定义一个信号量，通过以下命令导入到数据库中。
```sh
scalebox semaphore create --app-id=1 --sema-file my-sema.txt
```

#### 创建存储组配置文件

创建一个文本文件（如 `my-group.txt`），每行定义一个存储组和单位容量，供后续程序中使用：

```
"group0":20
"group1":10
```

### 2. 初始化 AggregatedPath

```go
import "beamform/internal/aggpath"

// 从文件创建 AggregatedPath
ap, err := aggpath.New(1, "my-group.txt")
if err != nil {
    // 处理错误
    log.Fatal(err)
}
```

### 3. 获取成员路径

```go
// 获取成员路径
memberPath, err := ap.GetMemberPath("group0", "task-123/data")
if err != nil {
    // 处理错误（如容量不足、存储组不存在等）
    log.Fatal(err)
}

// 使用分配的路径
fmt.Printf("分配的路径: %s\n", memberPath)
```

### 4. 释放成员路径

```go
// 释放成员路径
err = ap.ReleaseMemberPath("group0", "task-123/data")
if err != nil {
    // 处理错误
    log.Fatal(err)
}
```

## API 参考

### 类型

#### AggregatedPath

```go
type AggregatedPath struct {
    AppID           int
    StorageGroupMap map[string]int
}
```

- `AppID`: 应用ID，用于区分不同的应用实例
- `StorageGroupMap`: 存储组到容量的映射

### 函数

#### New

```go
func New(appID int, storageGroupFile string) (*AggregatedPath, error)
```

创建新的 `AggregatedPath` 实例。

- `appID`: 应用ID
- `storageGroupFile`: 存储组配置文件路径
- 返回值: `AggregatedPath` 实例或错误

#### GetMemberPath

```go
func (ap *AggregatedPath) GetMemberPath(storageGroup, path string) (string, error)
```

获取成员路径。

- `storageGroup`: 存储组标识
- `path`: 相对路径
- 返回值: 分配的成员路径或错误

#### ReleaseMemberPath

```go
func (ap *AggregatedPath) ReleaseMemberPath(storageGroup, path string) error
```

释放成员路径。

- `storageGroup`: 存储组标识
- `path`: 相对路径
- 返回值: 错误（如果释放失败）

## 配置

### 存储组配置文件格式

存储组配置文件是一个文本文件，每行定义一个存储组和容量：

```
"storage_group_name":capacity_gb
```

示例：
```
"group0":20
"group1":10
```

注意事项：
1. 存储组名可以用双引号括起来（支持带空格的名称）
2. 容量是整数，表示GB数
3. 空行和以 `#` 开头的行会被忽略

### 环境变量

- `PGHOST`: PostgreSQL数据库主机地址（用于variable和semaphore存储）

## 编程示例

### 完整示例

```go
package main

import (
    "fmt"
    "log"
    
    "beamform/internal/aggpath"
)

func main() {
    // 1. 创建 AggregatedPath
    ap, err := aggpath.New(1, "my-group.txt")
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 获取成员路径
    memberPath, err := ap.GetMemberPath("group0", "job-001/data")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("分配的路径: %s\n", memberPath)
    
    // 3. 使用路径...
    
    // 4. 释放路径
    err = ap.ReleaseMemberPath("group0", "job-001/data")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("路径已释放")
}
```

### 错误处理示例

```go
memberPath, err := ap.GetMemberPath("invalid-storage-group", "some/path")
if err != nil {
    if strings.Contains(err.Error(), "no valid storage group") {
        fmt.Println("错误: 存储组不存在")
    } else if strings.Contains(err.Error(), "No enough disk space") {
        fmt.Println("错误: 存储空间不足")
    } else {
        fmt.Printf("未知错误: %v\n", err)
    }
    return
}
```

## 测试

### 运行测试

```bash
cd apps/app-mwa/pipelines/beam-form/internal/aggpath
go test -v
```

### 测试覆盖

包包含以下测试：

1. **单元测试**:
   - `TestNew`: 测试创建功能，包括空文件路径、有效文件、不存在的文件等
   - `TestGetMemberPath`: 测试路径获取功能，包括基本功能、无效存储组、不同路径格式等
   - `TestReleaseMemberPath`: 测试路径释放功能，包括基本功能、无效存储组、不同路径格式等

2. **示例测试**:
   - `ExampleNew`: 展示如何使用New函数从文件创建AggregatedPath
   - `ExampleAggregatedPath_GetMemberPath`: 展示如何获取成员路径
   - `ExampleAggregatedPath_ReleaseMemberPath`: 展示如何释放成员路径

## 依赖

- `github.com/kaichao/scalebox/pkg/variable`: 变量存储
- `github.com/kaichao/scalebox/pkg/semagroup`: 信号量组管理
- `github.com/kaichao/scalebox/pkg/semaphore`: 信号量操作
- `github.com/kaichao/gopkg/errors`: 错误处理
- `github.com/sirupsen/logrus`: 日志记录

## 限制和注意事项

1. **并发安全**: 当前实现不是并发安全的，需要在调用层处理并发控制
2. **错误恢复**: 在分布式环境中，需要考虑网络分区和节点故障的恢复机制
3. **性能**: 频繁的数据库操作可能成为性能瓶颈，建议适当缓存

## 贡献

欢迎提交问题和拉取请求。在提交代码前，请确保：

1. 所有测试通过
2. 代码符合Go代码规范
3. 添加适当的测试用例
