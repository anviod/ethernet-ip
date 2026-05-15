# Ethernet/IP Go 库

Go 语言实现的 Ethernet/IP 协议库，支持与 Allen-Bradley PLC 等设备进行通信。

## 目录

- [功能特性](#功能特性)
- [支持的数据类型](#支持的数据类型)
- [架构设计](#架构设计)
- [调用流程](#调用流程)
- [API 参考](#api-参考)
- [示例代码](#示例代码)
- [性能优化](#性能优化)

---

## 功能特性

- ✅ TCP 连接管理（会话注册/注销）
- ✅ UDP 连接支持
- ✅ Tag 读写操作
- ✅ 批量 Tag 操作（TagGroup）
- ✅ 连接池支持（EIPTCPPool）
- ✅ UDT（用户定义类型）支持
- ✅ 符号寻址（Symbolic Addressing）
- ✅ 连接前向开放（Forward Open）
- ✅ 消息路由器（Message Router）
- ✅ 缓冲区池化优化
- ✅ 线程安全设计

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
│   └── bufferx.go       # 支持小端/大端读写、缓冲区池化
├── command/              # EIP 命令定义
│   └── command.go       # 命令常量（注册会话、发送数据等）
├── messages/             # 消息处理
│   ├── packet/          # 数据包编解码
│   ├── registerSession/ # 会话注册
│   ├── unRegisterSession/ # 会话注销
│   ├── listIdentity/    # 设备识别信息
│   ├── listInterface/   # 接口列表
│   ├── listServices/    # 服务列表
│   ├── sendRRData/      # 发送路由数据
│   └── sendUnitData/    # 发送单元数据
├── path/                # CIP 路径构建
│   └── path.go         # 逻辑路径、端口路径、数据路径
├── types/               # 类型定义
│   └── types.go        # 所有数据类型定义
├── utils/               # 工具函数
├── config.go            # 配置结构
├── context.go           # 上下文生成器
├── tcp.go              # TCP 连接管理
├── tcp_pool.go         # TCP 连接池
├── tag.go              # Tag 操作核心
├── request.go          # 请求处理
└── example.go          # 使用示例
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
┌─────────────────────────┐
│  NewTCPPool()           │  创建连接池
│  - 设置容量             │
│  - 配置工厂函数         │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  pool.Get()             │  获取连接
│  ┌─────────────────────┐│
│  │ 1. 从通道获取        ││  优先复用
│  │ 2. 通道为空则创建    ││
│  │ 3. 调用 Connect()   ││
│  └─────────────────────┘│
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  使用连接               │  执行业务操作
│  - Read/Write Tags      │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  pool.Put(conn)         │  归还连接
│  - 放回通道供复用       │
└──────┬──────────────────┘
       │
       ▼
   ┌───────┐
   │完成   │
   └───────┘
```

---

## API 参考

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

## 许可证

本项目采用 [WTFPL](http://www.wtfpl.net/) 许可证。

---

## 贡献

欢迎提交 Issue 和 Pull Request！
