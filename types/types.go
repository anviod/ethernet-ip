package types

// 基本整数类型
type Octet uint8
type Byte uint8

// 有符号整数类型
type SInt int8  // 8-bit signed integer
type Int int16  // 16-bit signed integer
type DInt int32 // 32-bit signed integer
type LInt int64 // 64-bit signed integer

// 无符号整数类型
type USInt uint8  // 8-bit unsigned integer
type UInt uint16  // 16-bit unsigned integer
type UDInt uint32 // 32-bit unsigned integer
type ULInt uint64 // 64-bit unsigned integer

// 别名类型 - 兼容常见命名规范
type INT8 int8
type UINT8 uint8
type INT16 int16
type UINT16 uint16
type INT32 int32
type UINT32 uint32
type INT64 int64
type UINT64 uint64

// 浮点类型
type Real float32  // 32-bit floating point
type LReal float64 // 64-bit floating point

// 浮点类型别名
type FLOAT float32
type DOUBLE float64

// 布尔类型
type Bool uint8

// 位类型
type BIT uint8

// 字符串类型
type STRING string

// 字类型 (Word types)
type Word uint16  // 16-bit word
type DWord uint32 // 32-bit double word
type LWord uint64 // 64-bit long word
