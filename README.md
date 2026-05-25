# Ethernet/IP Go 库

Go 语言实现的 Ethernet/IP 协议库，支持与 Allen-Bradley PLC 等设备进行通信。

## 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
  - [安装库](#安装库)
    - [使用最新版本](#使用最新版本)
    - [指定版本](#指定版本)
  - [导入库](#导入库)
  - [创建连接](#创建连接)
  - [读取 Tag](#读取-tag)
  - [写入 Tag](#写入-tag)
  - [批量操作](#批量操作)
- [支持的数据类型](#支持的数据类型)
- [架构设计](#架构设计)
- [调用流程](#调用流程)
- [访问模式](#访问模式)
  - [标准 CIP 模式](#标准-cip-模式)
  - [Logix 模式](#logix-模式)
- [配置选项](#配置选项)
- [诊断监控](#诊断监控)
- [UDP 设备发现](#udp-设备发现)
- [事务性支持](#事务性支持)
- [cpppo 兼容性](#cpppo-兼容性)
- [API 参考](#api-参考)
- [示例代码](#示例代码)
- [测试](#测试)
- [性能优化](#性能优化)
- [版本更新](#版本更新)

---

## 功能特性

- ✅ TCP 连接管理（会话注册/注销）
- ✅ UDP 连接支持
- ✅ UDP 设备发现（Broadcast Discovery）
- ✅ Tag 读写操作
- ✅ 批量 Tag 操作（TagGroup）
- ✅ 连接池支持（EIPTCPPool）
- ✅ UDT（用户定义类型）支持
- ✅ 符号寻址（Symbolic Addressing）
- ✅ 连接前向开放（Forward Open）/关闭（Forward Close）
- ✅ 消息路由器（Message Router）
- ✅ 缓冲区池化优化
- ✅ 线程安全设计
- ✅ 兼容（Logix Class 2 对象标签访问）
- ✅ 连接状态检查（IsConnected）
- ✅ 自动重连机制（指数退避，最多 3 次重试）
- ✅ 批量数据包操作（BatchRead/BatchWrite）
- ✅ 超时配置（连接超时、读写超时、操作超时）
- ✅ 事务性支持（TagGroup 原子写入）
- ✅ 诊断监控（连接状态、统计指标、事件监听）
- ✅ STRING 写入优化（单次请求完成）

---
## 快速开始

### 安装库

#### 使用最新版本

```bash
go get github.com/anviod/ethernet-ip@latest
```

#### 指定版本

```bash
# 使用特定版本（推荐用于生产环境）
go get github.com/anviod/ethernet-ip@v0.0.4

# 使用特定提交
go get github.com/anviod/ethernet-ip@abc1234

# 使用分支
go get github.com/anviod/ethernet-ip@main
```

### 导入库

```go
import (
    "log"
    "github.com/anviod/ethernet-ip"
)
```

### 创建连接

```go
// 创建 TCP 连接对象
conn, err := ethernet_ip.NewTCP("192.168.1.10", nil)
if err != nil {
    log.Fatal("创建连接失败:", err)
}
defer conn.Close()

// 建立连接
if err := conn.Connect(); err != nil {
    log.Fatal("连接失败:", err)
}
```

### 读取 Tag

```go
// 获取所有 Tag
tags, err := conn.AllTags()
if err != nil {
    log.Fatal("获取 Tag 列表失败:", err)
}

// 读取单个 Tag
tag := tags["Counter"]
if err := tag.Read(); err != nil {
    log.Fatal("读取 Tag 失败:", err)
}

// 获取值
log.Printf("Tag 值: %d\n", tag.Int32())
```

### 写入 Tag

```go
// 设置新值并写入
tag.SetInt32(12345)
if err := tag.Write(); err != nil {
    log.Fatal("写入 Tag 失败:", err)
}
log.Println("写入成功")
```

### 批量操作

```go
// 创建 Tag 组
lock := new(sync.Mutex)
group := ethernet_ip.NewTagGroup(lock)

// 添加多个 Tag
group.Add(tags["tag1"])
group.Add(tags["tag2"])

// 批量读取
if err := group.Read(); err != nil {
    log.Fatal("批量读取失败:", err)
}

// 批量写入
tags["tag1"].SetInt32(100)
tags["tag2"].SetString("updated")
if err := group.Write(); err != nil {
    log.Fatal("批量写入失败:", err)
}
```

---

## 支持的数据类型

| 类型 | 说明 | 符号 |
|------|------|------|
| `BOOL` | 布尔型 | `0xc1` |
| `SINT` / `INT8` | 8位有符号整数 | `0xc2` |
| `INT` / `INT16` | 16位有符号整数 | `0xc3` |
| `DINT` / `INT32` | 32位有符号整数 | `0xc4` |
| `LINT` / `INT64` | 64位有符号整数 | `0xc5` |
| `USINT` / `UINT8` | 8位无符号整数 | `0xc6` |
| `UINT` / `UINT16` | 16位无符号整数 | `0xc7` |
| `UDINT` / `UINT32` | 32位无符号整数 | `0xc8` |
| `ULINT` / `UINT64` | 64位无符号整数 | `0xc9` |
| `REAL` / `FLOAT` | 32位浮点数 | `0xca` |
| `LREAL` / `DOUBLE` | 64位浮点数 | `0xcb` |
| `STRING` | 字符串 | `0xfce` |

---

## 架构设计

### 模块结构

```
ethernet-ip/
├── bufferx/              # 字节缓冲区操作
│   ├── bufferx.go        # 支持小端/大端读写、缓冲区池化
│   ├── bufferx_test.go   # 单元测试
│   └── bufferx_benchmark_test.go # 性能测试
├── command/              # EIP 命令定义
│   ├── command.go        # 命令常量（注册会话、发送数据等）
│   └── command_test.go   # 单元测试
├── messages/             # 消息处理
│   ├── packet/          # 数据包编解码
│   │   ├── packet.go    # 数据包结构
│   │   ├── commonPacketFormat.go # CPF 格式
│   │   ├── messageRouter.go      # 消息路由器
│   │   ├── services.go  # 服务定义
│   │   ├── cmm.go       # CIP 消息管理
│   │   ├── ucmm.go      # 非连接消息管理
│   │   ├── data.go      # 数据项处理
│   │   ├── utils.go     # 工具函数
│   │   └── packet_test.go # 单元测试
│   ├── registerSession/  # 会话注册
│   ├── unRegisterSession/ # 会话注销
│   ├── listIdentity/     # 设备识别信息
│   ├── listInterface/    # 接口列表
│   ├── listServices/     # 服务列表
│   ├── sendRRData/       # 发送路由数据
│   ├── sendUnitData/     # 发送单元数据
│   └── nop/              # NOP 命令（空操作）
├── path/                # CIP 路径构建
│   ├── path.go          # 逻辑路径、端口路径、数据路径
│   └── path_test.go     # 单元测试
├── types/               # 类型定义
│   └── types.go         # 所有数据类型定义
├── utils/               # 工具函数
│   ├── len.go           # 长度计算
│   ├── mmap.go          # 内存映射
│   ├── mmap_unix.go     # Unix 平台内存映射
│   ├── mmap_windows.go  # Windows 平台内存映射
│   └── simd.go          # SIMD 优化
├── test/                # 集成测试
│   ├── cpppo/           # cpppo 兼容性测试
│   ├── protocol_verifier_test.go # 协议验证测试
│   ├── access_mode_test.go # 访问模式测试
│   ├── concurrency_test.go # 并发测试
│   └── benchmark_test.go # 性能基准测试
├── doc/                 # 文档
│   ├── PERFORMANCE_OPTIMIZATION*.md # 性能优化文档
│   ├── performance_report.json      # 性能报告
│   └── 2026年5月25日.md # 兼容性审查报告
├── config.go            # 配置结构（含超时配置）
├── context.go           # 上下文生成器
├── doc.go               # Go 文档注释
├── tcp.go               # TCP 连接管理（含重连机制、监控集成）
├── tcp_pool.go          # TCP 连接池
├── tag.go               # Tag 操作核心（含事务性支持）
├── request.go           # 请求处理（含 ForwardClose）
├── monitoring.go        # 诊断监控（连接状态、统计指标）
├── udp.go               # UDP 广播发现
├── example.go           # 使用示例
└── go.mod               # Go 模块配置
```

### 核心组件

```
┌─────────────────────────────────────────────────────────┐
│                    EIPTCP                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │  Config  │  │ TCPConn  │  │ Session  │              │
│  └──────────┘  └──────────┘  └──────────┘              │
│                                                         │
│  ┌──────────────────────────────────────────┐           │
│  │              Request Handler              │           │
│  │  - RegisterSession()                      │           │
│  │  - UnRegisterSession()                    │           │
│  │  - ListIdentity()                         │           │
│  │  - SendRRData() / SendUnitData()         │           │
│  └──────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                      Tag                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │   Name   │  │   Type   │  │  Value   │              │
│  └──────────┘  └──────────┘  └──────────┘              │
│                                                         │
│  ┌──────────────────────────────────────────┐           │
│  │           Read/Write Methods              │           │
│  │  - Read() / Write()                      │           │
│  │  - SetInt32() / SetString()              │           │
│  │  - Int32() / String()                    │           │
│  └──────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────┘
```

---

## 调用流程

### 1. 连接建立流程

```
┌─────────────┐
│   开始      │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  NewTCP()       │  创建 TCP 连接对象
│  - 解析地址     │
│  - 初始化配置   │
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│  Connect()      │  建立 TCP 连接
│  - DialTCP()    │
│  - SetKeepAlive │
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│ RegisterSession │  注册会话
│  - 生成上下文   │
│  - 发送注册请求 │
│  - 获取会话句柄 │
└──────┬──────────┘
       │
       ▼
   ┌───────┐
   │完成   │
   └───────┘
```

### 2. Tag 读取流程

```
┌─────────────┐
│   开始      │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│ AllTags()       │  获取所有 Tag
│  - 查询实例属性  │
│  - 构建 Tag 映射 │
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│  tag.Read()     │  读取单个 Tag
└──────┬──────────┘
       │
       ▼
┌─────────────────────────┐
│  TCP.Send()             │  发送请求
│  ┌─────────────────────┐│
│  │ readRequest()       ││  构建读请求
│  │ - 构建路径          ││
│  │ - 编码服务请求      ││
│  └─────────────────────┘│
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  TCP.SendRRData()       │  发送路由数据
│  - 封装为 CPF 包        │
│  - 发送到 PLC           │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  接收响应                │
│  - 解码 MessageRouter   │
│  - 解析响应数据         │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  readParser()           │  解析读取结果
│  - 提取类型            │
│  - 提取值              │
│  - 更新 Tag.value      │
└──────┬──────────────────┘
       │
       ▼
   ┌───────┐
   │完成   │
   └───────┘
```

### 3. Tag 写入流程

```
┌─────────────┐
│   开始      │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│ tag.SetInt32()  │  设置值（或其他 Set 方法）
│ - 序列化值      │
│ - 存入 wValue  │
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│  tag.Write()    │  执行写入
└──────┬──────────┘
       │
       ▼
┌─────────────────────────┐
│  writeRequest()         │  构建写请求
│  - 编码类型            │
│  - 编码值              │
│  - 构建路径            │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  TCP.Send()             │  发送写入请求
│  - multiple() 封装      │
│  - SendRRData/SendUnit │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  接收响应                │
│  - 检查状态             │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  更新 Tag.value         │
│  - copy(wValue, value) │
│  - 清空 wValue         │
└──────┬──────────────────┘
       │
       ▼
   ┌───────┐
   │完成   │
   └───────┘
```

### 4. 连接池使用流程

```
┌─────────────┐
│   开始      │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│ NewTCPPool()    │  创建连接池
│ - 指定容量     │
│ - 初始化连接   │
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│ Get()           │  获取连接
│ - 从池中获取   │
│ - 阻塞等待     │
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│  使用连接      │
│ - Tag 读写    │
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│ Put()           │  归还连接
│ - 放回池中     │
└──────┬──────────┘
       │
       ▼
   ┌───────┐
   │完成   │
   └───────┘
```

---

## 配置选项

### Config 结构

```go
type Config struct {
    TCPPort          uint16        // TCP 端口（默认 44818）
    UDPPort          uint16        // UDP 端口（默认 44818）
    Slot             uint8         // 控制器槽位
    TimeTick         types.USInt   // 连接时间刻度（毫秒）
    TimeTickOut      types.USInt   // 连接超时（TimeTick 单位）
    ConnectTimeout   time.Duration // 连接超时（默认 30 秒）
    ReadTimeout      time.Duration // 读取超时（默认 10 秒）
    WriteTimeout     time.Duration // 写入超时（默认 10 秒）
    OperationTimeout time.Duration // 操作超时（默认 15 秒）
}
```

### 使用示例

```go
config := ethernet_ip.DefaultConfig()
config.ConnectTimeout = time.Second * 60  // 自定义连接超时
config.ReadTimeout = time.Second * 15     // 自定义读取超时

conn, err := ethernet_ip.NewTCP("192.168.1.10", config)
```

---

## 诊断监控

### 连接状态

```go
state := conn.GetConnectionState()
fmt.Printf("当前状态: %s\n", state.String())
// 状态值: Disconnected, Connecting, Connected, Reconnecting
```

### 统计指标

```go
stats := conn.GetConnectionStats()
fmt.Printf("连接次数: %d\n", stats.ConnectCount)
fmt.Printf("断开次数: %d\n", stats.DisconnectCount)
fmt.Printf("重连次数: %d\n", stats.ReconnectCount)
fmt.Printf("总请求数: %d\n", stats.TotalRequests)
fmt.Printf("成功请求数: %d\n", stats.SuccessfulRequests)
fmt.Printf("失败请求数: %d\n", stats.FailedRequests)
fmt.Printf("平均响应时间: %v\n", stats.AvgResponseTime)
fmt.Printf("最小响应时间: %v\n", stats.MinResponseTime)
fmt.Printf("最大响应时间: %v\n", stats.MaxResponseTime)
fmt.Printf("总发送字节: %d\n", stats.TotalBytesSent)
fmt.Printf("总接收字节: %d\n", stats.TotalBytesReceived)
```

### 事件监听

```go
// 注册连接状态监听器
listener := func(event ethernet_ip.ConnectionEvent) {
    fmt.Printf("[%s] 状态变化: %s\n", 
        event.Timestamp.Format(time.RFC3339),
        event.State.String())
    
    if event.Error != nil {
        fmt.Printf("错误: %v\n", event.Error)
    }
    
    // 可以在这里实现告警、日志记录等逻辑
    if event.State == ethernet_ip.StateDisconnected {
        log.Warn("连接断开，准备重连...")
    } else if event.State == ethernet_ip.StateConnected {
        log.Info("连接已恢复")
    }
}

conn.AddConnectionListener(listener)

// 移除监听器
conn.RemoveConnectionListener(listener)
```

### 重置统计

```go
conn.ResetConnectionStats()
```

---

## UDP 设备发现

### 基本用法

```go
// 发现网络中的 EIP 设备（默认超时 5 秒）
devices, err := ethernet_ip.DiscoverDevices(time.Second * 5)
if err != nil {
    log.Fatal(err)
}

for _, device := range devices {
    fmt.Printf("设备名称: %s\n", device.DeviceName)
    fmt.Printf("IP地址: %s\n", device.IPAddress)
    fmt.Printf("MAC地址: %s\n", device.MacAddress)
    fmt.Printf("产品代码: %d\n", device.ProductCode)
    fmt.Printf("序列号: %d\n", device.SerialNumber)
    fmt.Printf("供应商ID: %d\n", device.VendorID)
    fmt.Printf("版本: %d.%d\n", device.RevisionMajor, device.RevisionMinor)
    fmt.Println("---")
}
```

### 指定端口

```go
// 使用自定义端口进行设备发现
devices, err := ethernet_ip.DiscoverDevicesWithPort(time.Second * 5, 44818)
```

### DiscoveredDevice 结构

| 字段 | 类型 | 说明 |
|------|------|------|
| IPAddress | net.IP | 设备 IP 地址 |
| MacAddress | string | MAC 地址 |
| DeviceName | string | 设备名称 |
| ProductCode | uint16 | 产品代码 |
| RevisionMajor | uint8 | 主版本号 |
| RevisionMinor | uint8 | 次版本号 |
| Status | uint16 | 设备状态 |
| SerialNumber | uint32 | 序列号 |
| VendorID | uint16 | 供应商 ID |

---

## 事务性支持

### TagGroup 原子写入

```go
// 创建 TagGroup 并启用事务性
group := ethernet_ip.NewTagGroup(nil)
group.AtomicWrite = true  // 启用原子写入模式

// 添加多个 Tag
group.Add(tag1)
group.Add(tag2)
group.Add(tag3)

// 设置新值
tag1.SetInt32(100)
tag2.SetString("test")
tag3.SetFloat32(3.14)

// 执行批量写入（事务性）
err := group.Write()
if err != nil {
    // 如果任何一个写入失败，所有标签的 changed 状态都会恢复
    log.Printf("写入失败，已回滚: %v\n", err)
} else {
    log.Println("所有标签写入成功")
}
```

### 事务性原理

当 `AtomicWrite` 启用时，`TagGroup.Write()` 会：

1. 收集所有待写入的标签
2. 发送批量写入请求
3. 检查每个响应的状态码
4. 如果任何一个写入失败，恢复所有标签的 `changed` 状态
5. 只有所有写入都成功，才更新标签的值

---

## 访问模式

本库支持两种访问模式来与 PLC 通信：**标准 CIP 模式**和 **Logix 模式**。

### 标准 CIP 模式

标准 CIP（Common Industrial Protocol）模式使用 Symbolic Addressing 方式访问标签。这种模式适用于大多数 Allen-Bradley PLC，包括 ControlLogix、CompactLogix 等。

**特点**：
- 使用符号地址（如 `Program:MainProgram.Counter`）
- 通过 Symbol Object (Class 0x6B) 访问标签
- 支持完整的标签路径解析
- 适用于标准 CIP 设备

**使用示例**：

```go
// 创建连接
conn, err := ethernet_ip.NewTCP("192.168.1.10", nil)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// 建立连接
if err := conn.Connect(); err != nil {
    log.Fatal(err)
}

// 使用标准 CIP 模式读取标签
tags, err := conn.AllTags()
if err != nil {
    log.Fatal(err)
}

tag := tags["Program:MainProgram.IntTag"]
if err := tag.Read(); err != nil {
    log.Fatal(err)
}
fmt.Printf("值: %d\n", tag.Int32())
```

**标准 CIP 模式调用流程**：

```
┌─────────────────────────────────────────────────────────────────┐
│                    标准 CIP 模式流程                            │
├─────────────────────────────────────────────────────────────────┤
│  1. AllTags()                                                  │
│     └─ 查询 Symbol Object (Class 0x6B)                         │
│         └─ 获取所有标签名称和类型                               │
│                                                                │
│  2. InitializeTag(name, tag)                                   │
│     └─ 构建符号路径                                             │
│         └─ 设置标签类型和连接引用                               │
│                                                                │
│  3. tag.Read()                                                 │
│     └─ SendRRData()                                            │
│         └─ CIP Read Tag Service (0x4C)                         │
│             └─ 返回 Message Router Response                    │
│                 └─ 解析响应数据到 tag.value                     │
│                                                                │
│  4. tag.Write()                                                │
│     └─ SendRRData()                                            │
│         └─ CIP Write Tag Service (0x4D)                        │
│             └─ 返回状态码                                       │
└─────────────────────────────────────────────────────────────────┘
```

### Logix 模式

Logix 模式使用 Logix Class 2 对象（Class 0x02）的属性访问方式。这种模式主要用于与 cpppo 服务器或某些特定的 Logix 设备通信。

**特点**：
- 使用属性 ID 访问（如属性 1 对应 BoolTag）
- 通过 Class 2, Instance 1 访问标签
- 使用 Get Attribute Single (0x0E) 服务
- 适用于 cpppo 模拟器和部分 Logix 设备

**使用示例**：

```go
// 创建连接
conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// 建立连接
if err := conn.Connect(); err != nil {
    log.Fatal(err)
}

// 使用 Logix 模式（Class 2 属性访问）读取标签
// 属性 ID 1 = BoolTag, 3 = IntTag, 10 = RealTag
data, err := conn.ReadClass2Attribute(3) // 读取 IntTag
if err != nil {
    log.Fatal(err)
}

// 解析 INT 类型数据
intValue := int16(binary.LittleEndian.Uint16(data[:2]))
fmt.Printf("IntTag 值: %d\n", intValue)
```

**Logix 模式调用流程**：

```
┌─────────────────────────────────────────────────────────────────┐
│                    Logix 模式流程                              │
├─────────────────────────────────────────────────────────────────┤
│  1. ReadClass2Attribute(attrID)                                │
│     └─ 构建 CIP 路径                                           │
│         └─ Class 2, Instance 1, Attribute attrID               │
│                                                                │
│  2. SendRRData()                                               │
│     └─ CIP Get Attribute Single (0x0E)                         │
│         └─ 返回 Message Router Response                        │
│             └─ 提取 ResponseData                               │
│                                                                │
│  3. 解析数据                                                   │
│     └─ 根据属性 ID 对应的数据类型解析                           │
│         └─ 返回原始字节数组供上层处理                           │
└─────────────────────────────────────────────────────────────────┘
```

### 模式对比

| 特性 | 标准 CIP 模式 | Logix 模式 |
|------|-------------|-----------|
| **访问方式** | Symbolic Addressing | Class 2 属性 |
| **地址格式** | `Program:MainProgram.Tag` | 属性 ID (1-12) |
| **适用设备** | 标准 Allen-Bradley PLC | cpppo 服务器、部分 Logix |
| **服务类型** | Read/Write Tag (0x4C/0x4D) | Get Attribute Single (0x0E) |
| **标签发现** | `AllTags()` 自动发现 | 需预先知道属性映射 |
| **UDT 支持** | ✅ | ❌ |

### 选择建议

- **生产环境**：使用 **标准 CIP 模式**，支持完整功能
- **测试环境**：使用 **Logix 模式**配合 cpppo 模拟器
- **兼容性测试**：两种模式都应测试

---

## cpppo 兼容性

本库已支持与 cpppo 服务器的通信。cpppo 服务器使用 Logix Class 2 对象来存储标签，与标准 CIP Symbolic Path 格式不同。

### 技术背景

cpppo 服务器将标签映射到 Logix Class 2, Instance 1 的属性，使用 CIP Get Attribute Single (0x0E) 服务进行访问：

| 标签名 | 属性 ID | 数据类型 |
|--------|---------|----------|
| `BoolTag` | 1 | BOOL |
| `SintTag` | 2 | SINT |
| `IntTag` | 3 | INT |
| `DintTag` | 4 | DINT |
| `LintTag` | 5 | LINT |
| `UsintTag` | 6 | USINT |
| `UintTag` | 7 | UINT |
| `UdintTag` | 8 | UDINT |
| `UlintTag` | 9 | ULINT |
| `RealTag` | 10 | REAL |
| `LrealTag` | 11 | LREAL |
| `StringTag` | 12 | STRING |

### 使用方法

使用 `ReadClass2Attribute` 和 `WriteClass2Attribute` 方法直接访问 Class 2 对象的属性：

```go
package main

import (
    "encoding/binary"
    "fmt"
    "log"
    "math"

    "github.com/anviod/ethernet-ip"
)

func main() {
    // 创建连接
    conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // 建立连接
    if err := conn.Connect(); err != nil {
        log.Fatal(err)
    }

    // 使用 Class 2 属性访问方式读取标签
    // 属性 ID 1 = BoolTag
    data, err := conn.ReadClass2Attribute(1)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("BoolTag 值: %v\n", data[0] != 0)

    // 写入 DINT 类型标签 (属性 ID 4 = DintTag)
    if err := conn.WriteClass2Attribute(4, []byte{0x39, 0x30, 0x00, 0x00}); err != nil {
        log.Fatal(err)
    }

    // 读取 REAL 类型标签
    data, err = conn.ReadClass2Attribute(10)
    if err != nil {
        log.Fatal(err)
    }
    floatValue := math.Float32frombits(binary.LittleEndian.Uint32(data[:4]))
    fmt.Printf("RealTag 值: %f\n", floatValue)
}
```

### 测试验证

运行 cpppo 兼容测试：

```bash
# 启动 cpppo 服务器
python test/cpppo/ethernet_ip_server_cpppo.py

# 运行 Go 测试
go test -v -run TestProtocolVerifier_Cpppo
```

测试结果示例：

```
[验证] Session 注册
✓ Session 注册成功: 0xC17173CA

[验证] Identity Object
✓ 供应商ID: 1
✓ 设备类型: 14
✓ 产品代码: 54

[验证] 数据类型支持 (cpppo Class 2 方式)
✓ BoolTag (BOOL): true
✓ SintTag (SINT): 42
✓ IntTag (INT): 12345
✓ DintTag (DINT): 987654321
✓ RealTag (REAL): 3.14159
✓ StringTag (STRING): Hello World

========================================
cpppo 兼容验证结果汇总: 通过=19, 失败=0
========================================
```

### 注意事项

1. cpppo 服务器需要先启动才能进行测试
2. 默认端口为 44818
3. 确保防火墙允许 TCP 44818 端口通信
4. 使用 Class 2 方式访问时，标签名需要与 cpppo 服务器配置的标签名一致

---

## API 参考

### 配置选项

#### Config 结构

```go
type Config struct {
    TCPPort          uint16        // TCP 端口（默认 44818）
    UDPPort          uint16        // UDP 端口（默认 44818）
    Slot             uint8         // 控制器槽位
    TimeTick         types.USInt   // 连接时间刻度（毫秒）
    TimeTickOut      types.USInt   // 连接超时（TimeTick 单位）
    ConnectTimeout   time.Duration // 连接超时（默认 30 秒）
    ReadTimeout      time.Duration // 读取超时（默认 10 秒）
    WriteTimeout     time.Duration // 写入超时（默认 10 秒）
    OperationTimeout time.Duration // 操作超时（默认 15 秒）
}
```

#### DefaultConfig() *Config

获取默认配置。

```go
config := DefaultConfig()
config.ConnectTimeout = time.Second * 60  // 自定义连接超时
conn, err := NewTCP("192.168.1.10", config)
```

### TCP 连接管理

#### NewTCP(address string, config *Config) (*EIPTCP, error)

创建新的 TCP 连接对象。

```go
conn, err := NewTCP("192.168.1.10", nil)
if err != nil {
    log.Fatal(err)
}
```

#### Connect() error

建立 TCP 连接并注册会话。

```go
err := conn.Connect()
if err != nil {
    log.Fatal(err)
}
defer conn.Close()
```

#### IsConnected() bool

检查连接是否已建立且会话已注册。

```go
if conn.IsConnected() {
    log.Println("连接已建立")
}
```

#### Close() error

关闭连接并注销会话。

```go
err := conn.Close()
```

### 会话管理

#### RegisterSession() error

注册 EIP 会话，获取会话句柄。

#### UnRegisterSession() error

注销当前会话。

#### ListIdentity() (*listIdentity.ListIdentity, error)

获取设备识别信息（设备类型、厂商、序列号等）。

```go
identity, err := conn.ListIdentity()
fmt.Println(identity)
```

#### ListInterface() (*listInterface.ListInterface, error)

获取网络接口信息。

#### ListServices() (*listServices.ListServices, error)

获取设备支持的服务列表。

### Tag 操作

#### AllTags() (map[string]*Tag, error)

获取所有可用的 Tag。

```go
tags, err := conn.AllTags()
for name, tag := range tags {
    fmt.Printf("Tag: %s, Type: %s\n", name, tag.TypeString())
}
```

#### InitializeTag(name string, tag *Tag)

初始化指定名称的 Tag。

```go
var tag Tag
conn.InitializeTag("OP.UDT_Alarm.DINT_065_096", &tag)
err := tag.Read()
```

### Tag 读取

#### Read() error

读取 Tag 的当前值。

```go
err := tag.Read()
if err != nil {
    log.Fatal(err)
}
value := tag.Int32()
```

#### GetValue() interface{}

获取 Tag 的当前值（自动适配类型）。

```go
value := tag.GetValue()
switch v := value.(type) {
case int32:
    fmt.Println("DINT:", v)
case string:
    fmt.Println("STRING:", v)
}
```

### Tag 写入

#### SetInt32(i int32)

设置 Tag 为 32 位整数。

```go
tag.SetInt32(123)
err := tag.Write()
```

#### SetString(s string)

设置 Tag 为字符串。

```go
tag.SetString("hello world")
err := tag.Write()
```

其他 Set 方法：
- `SetInt8(i int8)`
- `SetUInt8(i uint8)`
- `SetInt16(i int16)`
- `SetUInt16(i uint16)`
- `SetInt32(i int32)`
- `SetUInt32(i uint32)`
- `SetInt64(i int64)`
- `SetUInt64(i uint64)`
- `SetFloat32(f float32)`
- `SetFloat64(f float64)`

### Tag 值读取方法

| 方法 | 返回类型 | 说明 |
|------|---------|------|
| `Bool()` | `bool` | 获取布尔值 |
| `Int8()` | `int8` | 获取 8 位整数 |
| `UInt8()` | `uint8` | 获取 8 位无符号整数 |
| `Int16()` | `int16` | 获取 16 位整数 |
| `UInt16()` | `uint16` | 获取 16 位无符号整数 |
| `Int32()` | `int32` | 获取 32 位整数 |
| `UInt32()` | `uint32` | 获取 32 位无符号整数 |
| `Int64()` | `int64` | 获取 64 位整数 |
| `UInt64()` | `uint64` | 获取 64 位无符号整数 |
| `Float32()` | `float32` | 获取 32 位浮点数 |
| `Float64()` | `float64` | 获取 64 位浮点数 |
| `String()` | `string` | 获取字符串 |

### 批量操作 (TagGroup)

#### NewTagGroup(lock *sync.Mutex) *TagGroup

创建 Tag 组。

```go
lock := new(sync.Mutex)
group := NewTagGroup(lock)
```

#### Add(tag *Tag)

添加 Tag 到组。

```go
tag1 := tags["tag1"]
tag2 := tags["tag2"]
group.Add(tag1)
group.Add(tag2)
```

#### Read() error

批量读取所有 Tag。

```go
err := group.Read()
```

#### Write() error

批量写入所有已修改的 Tag。

```go
tag1.SetInt32(100)
tag2.SetString("updated")
err := group.Write()
```

### 批量数据包操作

#### BatchRead(count int) ([]*Packet, error)

从连接批量读取多个数据包。

```go
packets, err := conn.BatchRead(10)
if err != nil {
    log.Fatal(err)
}
for _, p := range packets {
    // 处理每个数据包
}
```

#### BatchWrite(packets []*Packet) error

批量写入多个数据包到连接。

```go
packets := []*packet.Packet{packet1, packet2, packet3}
err := conn.BatchWrite(packets)
if err != nil {
    log.Fatal(err)
}
```

### 诊断监控

#### GetConnectionState() ConnectionState

获取当前连接状态。

```go
state := conn.GetConnectionState()
fmt.Printf("连接状态: %s\n", state.String())
// 输出: Connected, Disconnected, Connecting, Reconnecting
```

#### GetConnectionStats() ConnectionStats

获取连接统计信息。

```go
stats := conn.GetConnectionStats()
fmt.Printf("连接次数: %d\n", stats.ConnectCount)
fmt.Printf("断开次数: %d\n", stats.DisconnectCount)
fmt.Printf("重连次数: %d\n", stats.ReconnectCount)
fmt.Printf("总请求数: %d\n", stats.TotalRequests)
fmt.Printf("成功请求数: %d\n", stats.SuccessfulRequests)
fmt.Printf("平均响应时间: %v\n", stats.AvgResponseTime)
```

#### AddConnectionListener(listener ConnectionEventListener)

注册连接状态变化监听器。

```go
listener := func(event ConnectionEvent) {
    fmt.Printf("状态变化: %s -> %s\n", 
        event.State.String(), 
        event.Timestamp.Format(time.RFC3339))
    if event.Error != nil {
        fmt.Printf("错误: %v\n", event.Error)
    }
}
conn.AddConnectionListener(listener)
```

#### ResetConnectionStats()

重置连接统计信息。

```go
conn.ResetConnectionStats()
```

### UDP 设备发现

#### DiscoverDevices(timeout time.Duration) ([]*DiscoveredDevice, error)

广播发现网络中的 EIP 设备。

```go
devices, err := ethernet_ip.DiscoverDevices(time.Second * 5)
if err != nil {
    log.Fatal(err)
}

for _, device := range devices {
    fmt.Printf("设备名称: %s\n", device.DeviceName)
    fmt.Printf("IP地址: %s\n", device.IPAddress)
    fmt.Printf("MAC地址: %s\n", device.MacAddress)
    fmt.Printf("产品代码: %d\n", device.ProductCode)
    fmt.Printf("序列号: %d\n", device.SerialNumber)
}
```

#### DiscoverDevicesWithPort(timeout time.Duration, port int) ([]*DiscoveredDevice, error)

指定端口进行设备发现。

```go
devices, err := ethernet_ip.DiscoverDevicesWithPort(time.Second * 5, 44818)
```

#### DiscoveredDevice 结构

```go
type DiscoveredDevice struct {
    IPAddress     net.IP   // 设备 IP 地址
    MacAddress    string   // MAC 地址
    DeviceName    string   // 设备名称
    ProductCode   uint16   // 产品代码
    RevisionMajor uint8    // 主版本号
    RevisionMinor uint8    // 次版本号
    Status        uint16   // 设备状态
    SerialNumber  uint32   // 序列号
    VendorID      uint16   // 供应商 ID
}
```

### TagGroup 事务性支持

#### AtomicWrite 字段

启用原子写入模式，确保批量写入的原子性。

```go
group := ethernet_ip.NewTagGroup(nil)
group.AtomicWrite = true  // 启用事务性写入

group.Add(tag1)
group.Add(tag2)

tag1.SetInt32(100)
tag2.SetInt32(200)

// 如果任何一个写入失败，所有标签的 changed 状态都会恢复
err := group.Write()
```

### 连接池

#### NewTCPPool(address string, config *Config, capacity int) (*EIPTCPPool, error)

创建连接池。

```go
pool, err := NewTCPPool("192.168.1.10", nil, 10)
if err != nil {
    log.Fatal(err)
}
defer pool.Close()
```

#### Get() (*EIPTCP, error)

从池中获取连接。

```go
conn, err := pool.Get()
if err != nil {
    log.Fatal(err)
}
defer pool.Put(conn)
```

#### Put(conn *EIPTCP) error

归还连接到池中。

#### Close() error

关闭连接池。

---

## 示例代码

### 基础使用

```go
package main

import (
    "log"

    "github.com/anviod/ethernet-ip"
)

func main() {
    // 创建连接
    conn, err := ethernet_ip.NewTCP("192.168.1.10", nil)
    if err != nil {
        log.Fatal("创建连接失败:", err)
    }

    // 建立连接
    err = conn.Connect()
    if err != nil {
        log.Fatal("连接失败:", err)
    }
    defer conn.Close()

    // 获取所有 Tag
    tags, err := conn.AllTags()
    if err != nil {
        log.Fatal("获取 Tag 列表失败:", err)
    }

    // 读取单个 Tag
    tag := tags["tagName"]
    err = tag.Read()
    if err != nil {
        log.Fatal("读取 Tag 失败:", err)
    }

    // 获取值
    log.Println("Tag 值:", tag.Int32())
}
```

### Tag 读取和写入

```go
package main

import (
    "log"

    "github.com/anviod/ethernet-ip"
)

func main() {
    conn, err := ethernet_ip.NewTCP("192.168.1.10", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    if err := conn.Connect(); err != nil {
        log.Fatal(err)
    }

    // 初始化 Tag
    var tag ethernet_ip.Tag
    conn.InitializeTag("Counter", &tag)

    // 读取
    if err := tag.Read(); err != nil {
        log.Fatal(err)
    }
    log.Printf("当前值: %d\n", tag.Int32())

    // 写入
    tag.SetInt32(12345)
    if err := tag.Write(); err != nil {
        log.Fatal(err)
    }
    log.Println("写入成功")
}
```

### 批量操作

```go
package main

import (
    "log"
    "sync"

    "github.com/anviod/ethernet-ip"
)

func main() {
    conn, err := ethernet_ip.NewTCP("192.168.1.10", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    if err := conn.Connect(); err != nil {
        log.Fatal(err)
    }

    tags, err := conn.AllTags()
    if err != nil {
        log.Fatal(err)
    }

    // 创建 Tag 组
    lock := new(sync.Mutex)
    group := ethernet_ip.NewTagGroup(lock)

    // 添加多个 Tag
    group.Add(tags["tag1"])
    group.Add(tags["tag2"])
    group.Add(tags["tag3"])

    // 批量读取
    if err := group.Read(); err != nil {
        log.Fatal(err)
    }

    // 设置新值
    tags["tag1"].SetInt32(100)
    tags["tag2"].SetInt32(200)
    tags["tag3"].SetString("updated")

    // 批量写入
    if err := group.Write(); err != nil {
        log.Fatal(err)
    }

    log.Println("批量操作完成")
}
```

### 使用连接池

```go
package main

import (
    "log"

    "github.com/anviod/ethernet-ip"
)

func main() {
    // 创建连接池
    pool, err := ethernet_ip.NewTCPPool("192.168.1.10", nil, 10)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // 获取连接
    conn, err := pool.Get()
    if err != nil {
        log.Fatal(err)
    }

    // 使用连接
    if err := conn.Connect(); err != nil {
        log.Fatal(err)
    }

    tags, err := conn.AllTags()
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("获取到 %d 个 Tag\n", len(tags))

    // 归还连接
    pool.Put(conn)
}
```

### 低级 API 使用

```go
package main

import (
    "log"

    "github.com/anviod/ethernet-ip"
)

func main() {
    conn, err := ethernet_ip.NewTCP("192.168.1.10", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    if err := conn.Connect(); err != nil {
        log.Fatal(err)
    }

    // 获取设备信息
    identities, err := conn.ListIdentity()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("设备识别: %+v\n", identities)

    // 获取接口信息
    interfaces, err := conn.ListInterface()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("接口信息: %+v\n", interfaces)

    // 获取服务列表
    services, err := conn.ListServices()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("服务列表: %+v\n", services)
}
```

### 读取不同数据类型

```go
package main

import (
    "fmt"
    "log"

    "github.com/anviod/ethernet-ip"
)

func main() {
    conn, err := ethernet_ip.NewTCP("192.168.1.10", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    if err := conn.Connect(); err != nil {
        log.Fatal(err)
    }

    tags, err := conn.AllTags()
    if err != nil {
        log.Fatal(err)
    }

    for name, tag := range tags {
        if err := tag.Read(); err != nil {
            continue
        }

        fmt.Printf("Tag: %s\n", name)
        fmt.Printf("  Type: %s\n", tag.TypeString())
        fmt.Printf("  Value: %v\n", tag.GetValue())
    }
}
```

---

## 测试

### 测试分类

本项目包含多种类型的测试，确保代码质量和功能正确性：

| 测试类型 | 描述 | 文件位置 |
|---------|------|---------|
| **单元测试** | 测试单个函数或方法的行为 | `*_test.go` |
| **集成测试** | 测试组件之间的交互 | `test/protocol_verifier_test.go` |
| **协议验证测试** | 验证 EIP 协议合规性 | `test/protocol_verifier_test.go` |
| **性能测试** | 测试性能指标 | `benchmark/` |

### 测试文件结构

```
ethernet-ip/
└── test/
    ├── protocol_verifier_test.go  # 协议验证集成测试
    ├── access_mode_test.go   # 访问模式测试（CIP/Logix）
```

### 测试覆盖范围

#### 1. 数据类型读写测试

支持所有 12 种数据类型的单独读写：

| 类型 | 写入测试 | 读取测试 | 验证方式 |
|------|---------|---------|---------|
| BOOL | `SetBool()` | `Bool()` | 布尔比较 |
| SINT | `SetInt8()` | `Int8()` | 数值比较 |
| INT | `SetInt16()` | `Int16()` | 数值比较 |
| DINT | `SetInt32()` | `Int32()` | 数值比较 |
| LINT | `SetInt64()` | `Int64()` | 数值比较 |
| USINT | `SetUInt8()` | `UInt8()` | 数值比较 |
| UINT | `SetUInt16()` | `UInt16()` | 数值比较 |
| UDINT | `SetUInt32()` | `UInt32()` | 数值比较 |
| ULINT | `SetUInt64()` | `UInt64()` | 数值比较 |
| REAL | `SetFloat32()` | `Float32()` | 精度比较 |
| LREAL | `SetFloat64()` | `Float64()` | 精度比较 |
| STRING | `SetString()` | `String()` | 字符串比较 |

#### 2. 批量操作测试

测试 TagGroup 的批量读写功能：

```go
// 批量读取测试
group := ethernet_ip.NewTagGroup(nil)
group.Add(tag1)
group.Add(tag2)
group.Read()  // 批量读取

// 批量写入测试
tag1.SetInt32(100)
tag2.SetFloat32(3.14)
group.Write() // 批量写入
```

#### 3. 访问模式测试

测试两种访问模式：

| 模式 | 测试函数 | 访问方式 |
|------|---------|---------|
| 标准 CIP 模式 | `TestAccessMode_CIPMode` | Symbolic Addressing |
| Logix 模式 | `TestAccessMode_LogixModeAttributeMapping` | Class 2 属性 |

### 运行测试

#### 运行所有测试

```bash
go test ./...
```

#### 运行特定测试

```bash
# 运行单元测试
go test -v -run "TestTag" ./...

# 运行协议验证测试
go test -v -run "TestProtocolVerifier" ./test/...

# 运行性能测试
go test -bench=. ./benchmark/...
```

#### 运行集成测试（需要 cpppo 服务器）

```bash
# 启动 cpppo 服务器
python test/cpppo/ethernet_ip_server_cpppo.py

# 在另一个终端运行测试
go test -v -run "TestProtocolVerifier_Cpppo" ./test/...
```

### 测试流程

#### 1. 单元测试流程

```
┌─────────────────────────────────────────────────────────┐
│                  单元测试流程                          │
├─────────────────────────────────────────────────────────┤
│  1. 初始化测试环境                                     │
│     └─ 创建 Tag 对象                                   │
│         └─ 设置测试数据                                │
│                                                       │
│  2. 执行测试方法                                       │
│     └─ 调用被测函数/方法                               │
│         └─ 捕获返回值                                 │
│                                                       │
│  3. 验证结果                                          │
│     └─ 使用 t.Errorf() / t.Fatalf()                   │
│         └─ 输出测试失败信息                           │
│                                                       │
│  4. 清理资源                                          │
│     └─ 使用 defer 释放资源                            │
└─────────────────────────────────────────────────────────┘
```

#### 2. 集成测试流程

```
┌─────────────────────────────────────────────────────────┐
│                  集成测试流程                          │
├─────────────────────────────────────────────────────────┤
│  1. 建立连接                                          │
│     └─ NewTCP() 创建连接对象                          │
│         └─ Connect() 建立 TCP 连接                   │
│             └─ RegisterSession() 注册会话             │
│                                                       │
│  2. 执行测试用例                                       │
│     ├─ Session 验证                                   │
│     ├─ Identity 查询                                  │
│     ├─ 数据类型读写验证                               │
│     ├─ Tag 单独读写                                   │
│     ├─ TagGroup 批量操作                              │
│     └─ 错误处理测试                                   │
│                                                       │
│  3. 收集结果                                          │
│     └─ 统计通过/失败数量                              │
│         └─ 输出测试报告                               │
│                                                       │
│  4. 清理连接                                          │
│     └─ Close() 关闭连接                               │
│         └─ UnRegisterSession() 注销会话              │
└─────────────────────────────────────────────────────────┘
```

#### 3. 协议验证测试流程

```
┌─────────────────────────────────────────────────────────┐
│                协议验证测试流程                        │
├─────────────────────────────────────────────────────────┤
│  1. 连接到测试设备                                    │
│     └─ 连接到 cpppo 模拟器或真实 PLC                  │
│                                                       │
│  2. 验证协议合规性                                     │
│     ├─ EIP TCP 协议握手                               │
│     ├─ CIP 路径构建                                   │
│     ├─ Message Router 服务                           │
│     ├─ Tag 读写服务 (0x4C/0x4D)                      │
│     └─ Class 2 属性访问 (0x0E)                        │
│                                                       │
│  3. 验证数据完整性                                    │
│     ├─ 字节序转换                                     │
│     ├─ 缓冲区操作                                     │
│     └─ 数据编码/解码                                  │
│                                                       │
│  4. 生成测试报告                                      │
│     └─ 输出 JSON/Markdown 格式报告                    │
│         └─ 保存到 doc/ 目录                           │
└─────────────────────────────────────────────────────────┘
```

### 测试结果示例

#### 单元测试结果

```
=== RUN   TestTag_SetInt32
--- PASS: TestTag_SetInt32 (0.00s)
=== RUN   TestTag_GetInt16
--- PASS: TestTag_GetInt16 (0.00s)
=== RUN   TestTag_GetString
--- PASS: TestTag_GetString (0.00s)
=== RUN   TestTag_Bool
--- PASS: TestTag_Bool (0.00s)
PASS
ok      github.com/anviod/ethernet-ip    0.001s
```

#### 集成测试结果

```
[验证] Session 注册
✓ Session 注册成功

[验证] Identity Object
✓ 供应商ID: 1
✓ 设备类型: 14
✓ 产品代码: 54
✓ 产品名称: cpppo
✓ 产品序列号: 0x12345678

[验证] 数据类型支持 (cpppo Class 2 方式)
✓ BoolTag (BOOL): true
✓ SintTag (SINT): 42
✓ IntTag (INT): 12345
✓ DintTag (DINT): 987654321
✓ RealTag (REAL): 3.14159
✓ StringTag (STRING): Hello World

========================================
验证结果汇总: 通过=19, 失败=0
========================================
```

### 测试报告

测试完成后会生成性能报告到 `doc/` 目录：

```bash
# 生成性能报告
go test -bench=. -benchmem ./benchmark/... > doc/BENCHMARK_RESULT.md
```

---

## 性能优化

本库已进行多项性能优化：

### 已实现的优化

1. **缓冲区池化** - 使用 `sync.Pool` 复用缓冲区，减少 GC 压力
2. **随机数生成器优化** - 预初始化种子，避免重复初始化开销
3. **TCP 缓冲区复用** - 连接级别的读取缓冲区复用
4. **错误短路机制** - 早期错误返回，减少无效操作
5. **预分配容量** - `NewWithCapacity()` 减少重新分配

### 性能基准

| 操作 | 吞吐量 | 内存分配 |
|-----|-------|---------|
| BufferPool_Reuse | ~8000万 ops/s | 2 B/op |
| Tag 读取 | ~300万 ops/s | varies |
| Tag 写入 | ~200万 ops/s | varies |

详细性能测试报告请参考 [doc/PERFORMANCE_OPTIMIZATION_CN.md](doc/PERFORMANCE_OPTIMIZATION_CN.md)

---

## 版本更新

### v0.0.6 (2026-05-25)

**新增功能**
- 新增 **STRING 写入优化** - 简化字符串写入逻辑，单次请求完成
- 新增 **超时配置** - 支持连接超时、读取超时、写入超时、操作超时
- 新增 **性能基准测试** - 添加完整的性能测试套件
- 新增 **事务性支持** - TagGroup 原子写入，失败自动回滚
- 新增 **诊断监控** - 连接状态监控、统计指标收集、事件监听
- 新增 **UDP 广播发现** - 设备自动发现功能
- 新增 **ForwardClose** - 连接关闭方法，避免资源泄漏
- 新增 **指数退避重连** - 优化重连机制，避免服务器压力

**Bug 修复**
- 修复字符串写入需要两次请求的问题
- 修复重连间隔为 0 的问题

**文档更新**
- 更新架构目录结构
- 添加新功能 API 文档
- 更新兼容性审查报告

### v0.0.5 (2026-05-22)

**Bug 修复**
- 修复 TCP 连接异常断开后的重连逻辑
- 修复 ReadFromFile/WriteToFile 方法未实现的占位错误处理
- 修复 BatchRead/BatchWrite 方法文档缺失

**新增功能**
- 新增 `IsConnected()` 方法，用于检查连接状态
- 新增重连机制（最多重试 3 次）
- 新增 `BatchRead()` 和 `BatchWrite()` 批量数据包操作方法
- 完善文档注释，补充 doc.go 完整 API 文档

**性能优化**
- 优化 TCP 缓冲区复用策略
- 优化连接池获取/归还逻辑

### v0.0.4 (2026-05-19)

**Bug 修复**
- 修复 NOP 命令空数据导致 cpppo 状态机错误
- 修复 Tag.Write() 缓冲区复制问题 (`copy` → `append`)
- 修复 `readParser` offset 计算错误（扩展类型 0x2a0 场景）
- 移除 Write() 函数冗余的 nil 检查
- 添加 offset 负值边界检查，避免切片越界 panic

**新增功能**
- 新增 Tag 单元测试（14 个测试函数）
- 新增性能测试套件（生成 JSON/Markdown 报告）

### v0.0.3 (2026-05-18)

**Bug 修复**
- 修复 TagGroup 批量写入失败问题

**新增功能**
- 添加 cpppo 服务器兼容性支持（Logix Class 2 对象标签访问）
- 新增 `ReadClass2Attribute` 方法

### v0.0.2 (2026-05-17)

**新增功能**
- 添加连接池支持（EIPTCPPool）
- 添加缓冲区池化优化
- 支持 UDT（用户定义类型）

### v0.0.1 (2026-05-16)

**初始版本**
- 基础 TCP 连接管理
- Tag 读写操作
- 支持基本数据类型（BOOL, INT, DINT, REAL, STRING 等）
- 消息路由器实现

---

## 许可证

本项目采用 [WTFPL](http://www.wtfpl.net/) 许可证。

---

## 贡献

欢迎提交 Issue 和 Pull Request！
