# vpath — 路径映射引擎

把物理上分散的多个存储空间组织成若干**逻辑存储类别**，对上层只暴露一个
路径查询接口。完整设计见 [DESIGN.md](DESIGN.md)。

> 上一版实现已在 [`internal/vpath0`](../vpath0/README.md) 归档保留（无生产调用方，仅供对照），
> 两者的取舍见 [DESIGN.md](DESIGN.md) 的「三版实现对比」。

## 三类存储虚拟化

| | a 聚合 | b1 复制 | b2 合并 |
|---|---|---|---|
| 类比 | JBOD / LVM | RAID1 | RAID0 |
| 选择时机 | 写（每次） | 读（每次） | 写（仅首次） |
| 选择依据 | 剩余量最大 | 配置权重 | 配置权重 |
| 状态 | 容量账本 + 位置记录 | **无** | 位置记录 |
| 对外容量 | Σ 成员 | 单空间 | Σ 成员 |

三类共用一个无状态加权选择实现（rendezvous hashing），同一 key 恒选同一
目标，多进程天然一致。

**类型不在配置里声明**，由调用方选择的 API 决定：同一份 `targets`，调
`Allocate` 即写一份，调 `AllocateAll` 即写多份。类型 a 的输出可作为组合层
（b1 / b2）的成员，嵌套最多 2 层。

## 配置

```yaml
# ── 空间层：物理存储如何组织 ─────────────────────────────
pools:
  small:                        # 类型 a：小空间聚合成大空间
    strategy: max-free
    # 成员不在配置里——见信号量组 vpath:free-gb:small

# ── 类别层：一类数据放在哪 ───────────────────────────────
classes:
  origin-tar:                   # 类型 b1：多个空间各存一份
    targets:
      - path: "astro@10.100.1.30:10022/data1/mydata"
        weight: 1
      - path: "astro@10.100.1.30:10022/data2/mydata"
        weight: 3

  cross-site:                   # 类型 b2：两个池容量相加
    unit_size_gb: 11
    targets:
      - pool: site-a  weight: 1
      - pool: site-b  weight: 1
```

## 池成员

**不在配置里声明**——池有哪些空间，完全由信号量组 `vpath:free-gb:<pool>`
决定。管理员用 sema 文件导入：

```sh
scalebox semaphore create --app-id=$app_id --sema-file preload.sema
```

每个成员一行，值就是它当前的剩余 GB 数。**信号量组既是容量账本，也是成员
清单**——只有这一份数据，vpath 只读不写。

`Load` 因此不查数据库，只读 YAML 建内存结构，与现有 vpath 一致，每个任务
进程的开销为零。成员路径从信号量名里切出（解析时注意路径自身可能含冒号）。

## API

```go
engine, _ := vpath.Load("/vpath.yaml", appID)

path, ok := engine.Locate(class, key)        // 只查不写，未分配时 ok=false
path, err := engine.Allocate(class, key)     // 确定位置，必要时新建
paths, err := engine.AllocateAll(class, key) // b1 写多份：全部副本路径
err := engine.Release(class, key)            // 按实际分配池释放
```

`Locate` 与 `Allocate` 分开，是为了避免「查一个尚未分配的东西，却凭空
分配一块空间并扣掉容量」。

容量一律取 class 的 `unit_size_gb`，不由调用方传入——避免分配与释放各传
一个 size 而账本对不上。`AllocateAll` 只对类型 b1 有意义，对以 pool 为主的
类别调用会造成 N 倍容量分配。

## 命名约定

| 键 | 含义 |
|---|---|
| `vpath:free-gb:<pool>:<member>` | 成员剩余量（信号量） |
| `vpath:member-path:<pool>:<key>` | 数据单元 → 所在空间（共享变量） |

成员路径自身可能含冒号（如 `astro@host:10022/...`），解析信号量名时必须用
`SplitN(name, ":", 4)[3]`，不能用 `Split`。

## 测试

池的逻辑可脱离 scalebox 单测——`store` 接口抽出了存储后端，测试注入
`memoryStore`（见 `store_test.go`），无需启动 server：

```go
s := newMemoryStore()
s.setSema("vpath:free-gb:test-pool:/pool/node1", 100)
engine, _ := vpath.LoadWithStore("vpath.yaml", s)
```
