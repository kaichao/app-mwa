# scalebox pkg 接口升级 & vtask 升级影响分析

> 分析日期：2026-08-04  
> scalebox 版本：`v0.0.0-20260804151623-3ae88cb36754`

## 一、背景

两个上游变更需要在本项目中适配：

1. **scalebox/pkg 接口实现升级**：各 pkg 包（semaphore、vtask、semagroup、variable、global 等）从直接读写 PostgreSQL/Redis 改为通过 gRPC 客户端调用 server 端（controld）
2. **vtask 功能升级**：vtask 新增原子化 bind/unbind、add-subtask 等高级操作，以 `scalebox/examples/vtask/` 为模版

## 二、scalebox/pkg 接口升级影响

### 2.1 结论：Go 层函数签名全部兼容，无需改调用代码

所有 pkg 包对外暴露的函数签名未发生变化，仅在内部实现上将数据库直连改为 `pkg/client` 单例 gRPC 客户端。

### 2.2 各包影响明细

| 包 | 被本项目调用的函数 | 签名变化 | 需要改代码？ |
|---|---|---|---|
| `semaphore` | `AddValue`, `Create`, `CreateSemaphores`, `AddMapValues`, `GetValue` | 无 | 否 |
| `vtask` | `CreateSemaphore`, `CreateSemaphores`, `AddSemaphoreValue`, `AddSemaphoreMapValues` | 无 | 否 |
| `semagroup` | `Decrement` | 无（但行为已变） | **是**（见第三节） |
| `variable` | `GetValue`, `Set` | 无 | 否 |
| `global` | `Get`, `Set` | 无 | 否 |
| `task` | `Add`, `AddWithMapHeaders`, `AddTasks`, `AddTasksWithMapHeaders` | 无（仍走 CLI，非 gRPC） | 否 |
| `common` | `AddTimeStamp`, `SetJSONAttribute`, `AppendToFile` | 无（纯工具函数） | 否 |

### 2.3 前置条件变化

- **`GRPC_SERVER` 环境变量**：所有 gRPC 化后的 pkg 函数通过 `GRPC_SERVER` 定位 controld（默认 `localhost:50051`）。当前 `app/app.yaml` 中 `router-beam-form` 模块已通过 `cluster-1.parameters.grpc_server` 配置该值，无需额外设置。
- **`go.mod` 间接依赖**：`google.golang.org/grpc v1.81.1` 和 `google.golang.org/protobuf v1.36.11` 已在 go.mod 间接依赖中，无需新增。
- **行为变化**：批量/正则操作从单条原子 SQL 变为多次 RPC 循环（如 `AddMapValues`），原子性降级，但本项目不依赖该场景的原子性。
- **不再需要 Redis / PostgreSQL 直连**：`pkg` 函数不再读取 `REDIS_HOST`、`PGHOST` 等环境变量。

### 2.4 可清理的遗留代码

`app/router/init.go:46`：
```go
os.Setenv("REDIS_HOST", os.Getenv("GRPC_SERVER"))
```
这行在 gRPC 体系下已无意义，可以删除。

## 三、vtask 功能升级影响

### 3.1 新旧架构差异

| 方面 | 旧模式（当前 beam-form） | 新模式（examples/vtask） |
|---|---|---|
| **资源绑定** | Go router 中 `semagroup.Decrement(":slot_vtask_size:vtask-head:")` | shell 脚本 `vtask bind`（原子化） |
| **子任务创建** | Go router 中 `task.AddWithMapHeaders` 到 vtask-head | shell 脚本 `vtask add-subtask --direct`（同步）/ 异步路径 |
| **gate 释放** | Go router 中 `semaphore.AddValue("vtask_size:wait-queue", 1)` | shell 脚本 `from-vtask-head.sh` |
| **资源释放** | Go router 中 `vtask.AddSemaphoreValue(":"+sema, 1, 0, appID)` | shell 脚本 `from-vtask-tail.sh` 调用 `vtask unbind` |
| **双信号量** | 无明确区分 | 流控版（controld 自动管理）+ 可编程版（脚本操作） |

### 3.2 需要修改的代码

#### 3.2.1 `app/router/vtask_head.go` — `toVtaskHead()`（第 62-84 行）

**当前代码**：
```go
func toVtaskHead(cubeName string) error {
    groupName := ":slot_vtask_size:vtask-head:"
    semaName, _, err := semagroup.Decrement(groupName, appID)
    slotSeq, _ := strconv.Atoi(semaName[len(groupName):])
    // ...
    headers := map[string]string{
        "_vtask_cube_name": cubeName,
        "to_slot_index":    fmt.Sprintf("%d", slotSeq),
        "_slot_seq":        fmt.Sprintf("%d", slotSeq),
    }
    // task.AddWithMapHeaders(cubeName, headers, envs)
}
```

**改为**：
```go
func toVtaskHead(cubeName string) error {
    resource, mode, err := vtask.BindResource(appID, "")
    if err != nil {
        return errors.WrapE(err, "vtask-bind", "cube-name", cubeName)
    }
    // resource: slot_seq for GROUP-BOUND (SLOT-BOUND), hostname for HOST-BOUND
    // mode: "SLOT-BOUND" / "HOST-BOUND" / "DEFAULT"
    // ...
}
```

**变更要点**：
- `semagroup.Decrement` 替换为 `vtask.BindResource(appID, "")`
- `BindResource` 返回 `(resource string, mode string, err error)`
- GROUP-BOUND 下 resource 为 slot_seq 字符串（如 `"0"`），HOST-BOUND 下为 hostname（如 `"n0-0"`）
- 不再需要手动解析信号量名提取 slotSeq
- `import` 中移除 `"github.com/kaichao/scalebox/pkg/semagroup"`

#### 3.2.2 `app/router/vtask_tail.go` — `fromVtaskTail()`（第 11-22 行，可选改）

**当前代码**：
```go
func fromVtaskTail(body string, headers map[string]string) error {
    semaName := ":" + headers["_vtask_size_sema"]
    _, err := vtask.AddSemaphoreValue(semaName, 1, vtaskID, appID)
    // ...
}
```

**可改为**（语义更清晰）：
```go
func fromVtaskTail(body string, headers map[string]string) error {
    vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
    semaName := ":" + headers["_vtask_size_sema"]
    err := vtask.UnbindResource(appID, vtaskID, semaName)
    // ...
}
```

**注意**：`vtask.UnbindResource` 要求 vtaskID 和 semaName 至少提供一个。当前代码中 vtaskID 被硬编码为 0，需确认是否从 headers 中正确解析。

#### 3.2.3 `app/router/init.go` — 清理遗留代码

删除第 46 行：
```go
os.Setenv("REDIS_HOST", os.Getenv("GRPC_SERVER"))
```

### 3.3 不需要修改的部分

- **`fromVtaskHead()`**：gate 恢复（`semaphore.AddValue("vtask_size:wait-queue", 1, appID)`）和信号量创建（`vtask.CreateSemaphore`）逻辑仍有效。新模式中将 gate 释放移到 shell 脚本是架构选择，当前 Go router 中保留也可以工作。
- **`fromWaitQueue()`**：当前仅做转发 `toVtaskHead()`，行为不变。
- **`app/app.yaml`**：vtask 相关配置（`vtask_role: head/core/tail`, `vtask_size`, `vtask_size_sema_copy`, `task_dist_mode`）与新模型一致，无需修改。
- **所有 core 阶段函数**（`pull_unpack.go`, `beam_make.go`, `fits_redist.go`）：使用的 `vtask.AddSemaphoreValue`、`vtask.CreateSemaphores` 等函数签名不变。
- **`tar_load.go`**：使用的 `semaphore.AddValue`、`semaphore.CreateSemaphores` 签名不变，且 tar-load 在 vtask 管道之外独立运行。

### 3.4 新增可用的 vtask 函数

新模式在 `pkg/vtask` 中新增了以下函数，供未来使用：

| 函数 | 用途 |
|---|---|
| `vtask.BindResource(appID, moduleName)` | 原子化绑定计算资源 |
| `vtask.UnbindResource(appID, vtaskID, semaName)` | 释放计算资源 |
| `vtask.AddSubtask(appID, module, body, headers)` | 创建 vtask 子任务，建立父子关系 |
| `vtask.GetInfo(appID, vtaskID)` | 查询 vtask 详情 |
| `vtask.Fail(appID, vtaskID)` | 强制终止 vtask |
| `vtask.List(appID)` | 列出 app 下所有 vtask |
| `vtask.ListSubtasks(appID, vtaskID)` | 列出 vtask 子任务 |

## 四、改动汇总

| # | 文件 | 行号 | 改动类型 | 优先级 |
|---|---|---|---|---|
| 1 | `app/router/init.go` | 46 | 删除 `os.Setenv("REDIS_HOST", ...)` | 高 |
| 2 | `app/router/vtask_head.go` | 62-84 | `semagroup.Decrement` → `vtask.BindResource` | 高 |
| 3 | `app/router/vtask_head.go` | 1-9 | import 移除 `semagroup`，确认 `vtask` 已引入 | 高 |
| 4 | `app/router/vtask_tail.go` | 11-22 | 可选改 `vtask.AddSemaphoreValue` → `vtask.UnbindResource` | 低 |

## 五、不影响的部分（确认清单）

- ✅ `task.Add` / `task.AddWithMapHeaders` / `task.AddTasks` — 签名不变
- ✅ `vtask.CreateSemaphore` / `vtask.CreateSemaphores` — 签名不变
- ✅ `vtask.AddSemaphoreValue` / `vtask.AddSemaphoreMapValues` — 签名不变
- ✅ `semaphore.AddValue` / `semaphore.Create` / `semaphore.CreateSemaphores` — 签名不变
- ✅ `variable.GetValue` / `variable.Set` — 签名不变
- ✅ `global.Get` / `global.Set` — 签名不变
- ✅ `common.AddTimeStamp` / `common.SetJSONAttribute` — 签名不变
- ✅ `app/app.yaml` — vtask 配置已匹配新模型
- ✅ `go.mod` — gRPC 依赖已存在
- ✅ `app-base/` — 不使用 vtask，完全不受影响
- ✅ `internal/` 下所有包 — 不直接依赖 scalebox/pkg 接口变更部分
