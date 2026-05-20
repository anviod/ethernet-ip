/*
Package ethernet_ip 实现了一个 EtherNet/IP 客户端库，用于与 Allen-Bradley PLC 和兼容设备进行通信。

# 概述

EtherNet/IP 是 Rockwell Automation 开发的工业通信协议，用于制造和自动化环境。
这个库使 Go 应用程序能够连接到支持 EtherNet/IP 协议的 PLC（如 ControlLogix、CompactLogix 和 SLC 500 系列控制器）并交换数据。

该库提供以下功能：

  - TCP 连接管理及会话注册
  - 基于标签的读写操作，自动类型处理
  - 支持所有标准 CIP 数据类型（BOOL、SINT、INT、DINT、LINT、REAL 等）
  - 高性能场景的连接池支持
  - 通过 Logix Class 2 对象访问兼容 cpppo 服务器
  - 线程安全操作，带互斥锁保护
  - 符号地址解析，支持复杂标签路径

# 安装

使用以下命令安装包：

	go get github.com/anviod/ethernet-ip

# 基本用法

## 创建连接

创建到 PLC 的 TCP 连接：

	conn, err := ethernet_ip.NewTCP("192.168.1.10", nil)
	if err != nil {
	    log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.Connect(); err != nil {
	    log.Fatal(err)
	}

第二个参数是可选的 Config 结构体。传递 nil 使用默认设置。

## 读取标签

库支持多种读取标签的方式：

方法 1：读取所有标签并访问单个标签

	tags, err := conn.AllTags()
	if err != nil {
	    log.Fatal(err)
	}

	tag := tags["MyTag"]
	if err := tag.Read(); err != nil {
	    log.Fatal(err)
	}

	value := tag.Int32()
	log.Printf("标签值: %d", value)

方法 2：通过路径初始化特定标签

	tag := new(ethernet_ip.Tag)
	if err := conn.InitializeTag("Program:MainProgram.MyTag", tag); err != nil {
	    log.Fatal(err)
	}

	if err := tag.Read(); err != nil {
	    log.Fatal(err)
	}

## 写入标签

要写入标签值，使用 setter 方法后调用 Write：

	tag.SetInt32(42)
	if err := tag.Write(); err != nil {
	    log.Fatal(err)
	}

可用的 setter 方法：
  - SetBool(value bool)
  - SetInt8(value int8)
  - SetUInt8(value uint8)
  - SetInt16(value int16)
  - SetUInt16(value uint16)
  - SetInt32(value int32)
  - SetUInt32(value uint32)
  - SetInt64(value int64)
  - SetUInt64(value uint64)
  - SetFloat32(value float32)
  - SetFloat64(value float64)
  - SetString(value string)

## 读取值

成功读取后，使用适当的类型转换方法：

	tag.Read()

	// 数值类型
	intVal := tag.Int32()
	uintVal := tag.UInt16()
	floatVal := tag.Float32()

	// 字符串
	strVal := tag.String()

	// 布尔值
	boolVal := tag.Bool()

## 标签路径

可以使用完整的符号路径引用标签：

	// 简单标签
	tag := new(Tag)
	conn.InitializeTag("MyTag", tag)

	// 程序作用域中的标签
	conn.InitializeTag("Program:MainProgram.MyTag", tag)

	// 数组元素
	conn.InitializeTag("MyArray[0]", tag)

	// 多维数组
	conn.InitializeTag("MyArray[1,0,2]", tag)

	// 标签成员（UDT）
	conn.InitializeTag("MyUDT.MemberName", tag)

	// 嵌套 UDT 成员
	conn.InitializeTag("ParentUDT.ChildUDT.Member", tag)

## 使用标签组

标签组允许同时读写多个标签：

	lock := new(sync.Mutex)
	group := ethernet_ip.NewTagGroup(lock)

	tag1 := tags["Tag1"]
	tag2 := tags["Tag2"]
	group.Add(tag1)
	group.Add(tag2)

	// 读取多个标签
	if err := group.Read(); err != nil {
	    log.Fatal(err)
	}

	// 写入多个标签
	tag1.SetInt32(100)
	tag2.SetString("hello")
	if err := group.Write(); err != nil {
	    log.Fatal(err)
	}

## 连接池

对于高性能场景，使用连接池管理多个连接：

	pool, err := ethernet_ip.NewTCPPool("192.168.1.10", nil, 10)
	if err != nil {
	    log.Fatal(err)
	}
	defer pool.Close()

	// 从池中获取连接
	conn, err := pool.Get()
	if err != nil {
	    log.Fatal(err)
	}

	// 使用连接
	tags, err := conn.AllTags()
	// ...

	// 将连接放回池中
	pool.Put(conn)

## cpppo 服务器兼容性

库支持通过 Logix Class 2 对象属性访问标签，这是 cpppo 服务器使用的方式。
这允许与模拟 PLC 行为的软件模拟器通信。

	// 读取 Class 2 属性（属性 ID 1 对应 BoolTag）
	data, err := conn.ReadClass2Attribute(1)
	if err != nil {
	    log.Fatal(err)
	}

	// 写入 Class 2 属性（属性 ID 4 对应 DintTag）
	err = conn.WriteClass2Attribute(4, []byte{0x39, 0x30, 0x00, 0x00})
	if err != nil {
	    log.Fatal(err)
	}

## 设备发现

列出网络上设备的身份信息：

	identities, err := conn.ListIdentity()
	if err != nil {
	    log.Fatal(err)
	}
	for _, identity := range identities {
	    log.Printf("设备: %s, 类型: %d, 供应商: %d",
	        identity.ProductName, identity.DeviceType, identity.VendorID)
	}

## Forward Open

对于时间关键的通信，使用 Forward Open 建立专用连接路径：

	if err := conn.ForwardOpen(); err != nil {
	    log.Fatal(err)
	}
	defer conn.ForwardClose()

## 配置

Config 结构体允许自定义连接参数：

	config := &ethernet_ip.Config{
	    TCPPort:     44818,      // 默认 EtherNet/IP 端口
	    UDPPort:     44818,      // 默认 UDP 端口
	    Slot:        0,           // PLC 机架中的控制器槽号
	    TimeTick:    3,           // 时间刻度（毫秒）
	    TimeTickOut: 250,         // 连接超时（时间刻度单位）
	}

	conn, err := ethernet_ip.NewTCP("192.168.1.10", config)

# 数据类型

库支持所有标准 CIP 数据类型：

	CIP 类型 | Go 类型   | 大小（字节） | 描述
	----------|-----------|-------------|-------------------------------------
	BOOL      | bool      | 1           | 布尔值
	SINT      | int8      | 1           | 有符号 8 位整数
	INT       | int16     | 2           | 有符号 16 位整数
	DINT      | int32     | 4           | 有符号 32 位整数
	LINT      | int64     | 8           | 有符号 64 位整数
	USINT     | uint8     | 1           | 无符号 8 位整数
	UINT      | uint16    | 2           | 无符号 16 位整数
	UDINT     | uint32    | 4           | 无符号 32 位整数
	ULINT     | uint64    | 8           | 无符号 64 位整数
	REAL      | float32   | 4           | 单精度浮点数
	LREAL     | float64   | 8           | 双精度浮点数
	STRING    | string    | 可变        | CIP 字符串（最多 88 字节）
	STRING2   | string    | 可变        | CIP 字符串（扩展格式）

# 错误处理

所有方法都返回应该适当处理的错误：

	if err := tag.Read(); err != nil {
	    switch {
	    case errors.Is(err, ethernet_ip.ErrBufferTooShort):
	        log.Println("缓冲区太小，无法容纳标签数据")
	    case errors.Is(err, ethernet_ip.ErrTagNotFound):
	        log.Println("设备上不存在该标签")
	    default:
	        log.Printf("读取失败: %v", err)
	    }
	    return
	}

# 线程安全

库设计为线程安全：

  - 每个 Tag 有自己的互斥锁（Tag.Lock）保护读写操作
  - EIPTCP 有请求锁保护并发请求
  - 连接池使用互斥锁保护内部状态

您可以安全地并发使用多个标签：

	var wg sync.WaitGroup
	for _, t := range tags {
	    wg.Add(1)
	    go func(tag *ethernet_ip.Tag) {
	        defer wg.Done()
	        tag.Read()
	    }(t)
	}
	wg.Wait()

# 性能考虑

对于高性能场景：

 1. 使用连接池（EIPTCPPool）减少连接开销
 2. 使用标签组进行批量读写操作
 3. 对于连续监控，实现缓存而非轮询
 4. 考虑使用 Forward Open 进行时间关键的应用

# 常见问题

1. 连接拒绝：确保 PLC 可达，且 EtherNet/IP 端口未被防火墙阻止。

2. 标签未找到：验证标签路径是否正确。标签名称区分大小写。

3. 权限拒绝：某些 PLC 需要适当的访问级别才能读写标签。

4. 会话过期：如果会话过期，使用 Connect() 重新连接。

# 示例

请查看 examples 目录获取完整的工作示例：

  - 基本读写操作
  - 标签组操作
  - 连接池
  - cpppo 服务器通信

# 参考资料

有关 EtherNet/IP 协议的更多信息，请参阅：
  - https://www.rockwellautomation.com/en-us/technologies/industrial-protocols/ethernet-ip.html
  - https://www.odva.org/ethernet-ip

cpppo 服务器实现：
  - https://github.com/pjkundert/cpppo
*/
package ethernet_ip
