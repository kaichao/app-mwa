# vpath - 虚拟路径管理包

vpath包提供了一个虚拟路径管理功能，用于管理和分配文件系统路径，支持加权路径和聚合路径。加权路径可以根据权重进行智能分配管理；聚合路径主要用于在分布式系统中管理共享存储路径的分配和释放，合并多个共享目录下的存储配额，支持更大的存储容量需求，确保大型应用可共享有限的存储资源。

## 主要功能

- **加权路径管理**：直接使用预定义的路径
- **聚合路径管理**：从多个候选路径中动态分配
- **加权选择**：根据权重智能选择路径
- **路径释放**：使用后释放路径资源
- **多种聚合器**：支持memory和scalebox聚合器

## 快速开始

### 1. 从YAML配置文件创建

```yaml
# config.yaml
my-category:
  aggregator_type: "memory"
  weighted_paths:
    - path: "/fast/ssd"
      weight: 1.0
    - path: "/slow/hdd" 
      weight: 2.0
    - path: AGG_PATH
      weight: 3.0
      category: my-category
      need_gb: 20
  aggregated_paths:
    - name: my-category
      capacity_gb: 1000
      members:
        - "/agg/dir1"
        - "/agg/dir2"
```

```go
package main

import (
    "fmt"
    "beamform/internal/vpath"
)

func main() {
    // 创建VirtualPath实例
    vp, err := vpath.NewVirtualPath(1, "config.yaml")
    if err != nil {
        panic(err)
    }

    // 获取路径
    path, err := vp.GetPath("my-category", "job-001")
    if err != nil {
        panic(err)
    }
    fmt.Printf("分配到路径: %s\n", path)

    // 释放路径
    err = vp.ReleasePath("my-category", "job-001")
    if err != nil {
        panic(err)
    }
}
```

### 2. 编程方式创建

```go
package main

import (
    "fmt"
    "beamform/internal/vpath"
)

func main() {
    // 创建配置
    config := &vpath.Config{
        Name: "example",
        WeightedPaths: []vpath.WeightedPathConfig{
            {Path: "/path1", Weight: 1.0, Type: "static", Category: "default"},
            {Path: "/path2", Weight: 2.0, Type: "static", Category: "default"},
        },
        AggregatorType: "memory",
    }

    // 创建VirtualPath
    vp, err := vpath.NewVirtualPathFromConfig(1, config)
    if err != nil {
        panic(err)
    }

    // 使用示例
    for i := 0; i < 5; i++ {
        key := fmt.Sprintf("task-%d", i)
        path, err := vp.GetPath("default", key)
        if err != nil {
            panic(err)
        }
        fmt.Printf("%s -> %s\n", key, path)
    }
}
```

## 核心API

### NewVirtualPath(appID int, configFile string) (*VirtualPath, error)
从YAML配置文件创建VirtualPath实例。

### NewVirtualPathFromConfig(appID int, config *Config) (*VirtualPath, error)
从Config结构体创建VirtualPath实例。

### GetPath(category, key string) (string, error)
获取路径。相同category和key总是返回相同路径。

### ReleasePath(category, key string) error
释放路径，允许重新分配。

## 配置说明

### WeightedPathConfig
- `Path`: 路径（AGG_PATH表示聚合路径）
- `Weight`: 权重，影响选择概率
- `Type`: "static"或"aggregated"
- `Category`: 路径分类
- `CapacityGB`: 容量（GB）

### AggregatedPathConfig  
- `Name`: 聚合路径名称
- `CapacityGB`: 总容量
- `Members`: 成员路径列表

## 算法特性

vpath使用智能加权选择算法：
1. 选择实际占比/理论占比最小的项
2. 第一次调用选择权重最大的项
3. 权重≤0的项自动排除
4. 保证长期选择比例接近理论权重

## 示例场景

### 场景1：数据存储选择
```go
// 根据存储性能选择路径
path, _ := vp.GetPath("storage", "data-job")
// 可能返回："/fast/ssd" 或 "/slow/hdd"
```

### 场景2：计算节点分配  
```go
// 分配计算节点
node, _ := vp.GetPath("compute", "task-001")
// 可能返回："/node1" 或 "/node2" 或 "/node3"
```

### 场景3：临时工作目录
```go
// 获取临时工作目录
workdir, _ := vp.GetPath("temp", "process-001")
// 使用后释放
defer vp.ReleasePath("temp", "process-001")
```

## 注意事项

1. **key的唯一性**：相同category和key总是返回相同路径
2. **及时释放**：使用完路径后调用ReleasePath释放资源
3. **权重设置**：权重为0的路径不会被选择
4. **容量管理**：聚合路径有容量限制，分配时检查容量

## 更多示例

查看源码，获取完整示例。
