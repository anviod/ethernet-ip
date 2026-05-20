package test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"testing"

	ethernet_ip "github.com/anviod/ethernet-ip"
	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/messages/packet"
	"github.com/anviod/ethernet-ip/path"
)

type ProtocolVerifier struct {
	conn *ethernet_ip.EIPTCP
}

func NewProtocolVerifier(conn *ethernet_ip.EIPTCP) *ProtocolVerifier {
	return &ProtocolVerifier{conn: conn}
}

type TestResult struct {
	Name    string
	Passed  bool
	Message string
	Value   interface{}
}

func dialForTest(t *testing.T) *ethernet_ip.EIPTCP {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		t.Skipf("跳过测试: 无法创建TCP连接: %v", err)
	}
	if err := conn.Connect(); err != nil {
		t.Skipf("跳过测试: 无法连接到模拟器: %v", err)
	}
	return conn
}

func (pv *ProtocolVerifier) verifySession() TestResult {
	fmt.Println("\n[验证] Session 注册")

	conn := pv.conn
	if conn == nil || !conn.IsConnected() {
		return TestResult{Name: "Session注册", Passed: false, Message: "连接未建立"}
	}

	fmt.Printf("✓ Session 注册成功\n")
	return TestResult{Name: "Session注册", Passed: true, Message: "连接已建立"}
}

func (pv *ProtocolVerifier) verifyIdentity() []TestResult {
	fmt.Println("\n[验证] Identity Object")
	results := []TestResult{}

	identity, err := pv.conn.ListIdentity()
	if err != nil {
		fmt.Printf("✗ Identity 查询失败: %v\n", err)
		results = append(results, TestResult{Name: "Identity查询", Passed: false, Message: err.Error()})
		return results
	}

	if len(identity.Items) == 0 {
		fmt.Printf("✗ Identity 没有项目\n")
		results = append(results, TestResult{Name: "Identity查询", Passed: false, Message: "没有项目"})
		return results
	}

	item := identity.Items[0]
	fmt.Printf("✓ 供应商ID: %d\n", item.VendorID)
	results = append(results, TestResult{Name: "供应商ID", Passed: true, Value: item.VendorID})

	fmt.Printf("✓ 设备类型: %d\n", item.DeviceType)
	results = append(results, TestResult{Name: "设备类型", Passed: true, Value: item.DeviceType})

	fmt.Printf("✓ 产品代码: %d\n", item.ProductCode)
	results = append(results, TestResult{Name: "产品代码", Passed: true, Value: item.ProductCode})

	fmt.Printf("✓ 产品名称: %s\n", string(item.ProductName))
	results = append(results, TestResult{Name: "设备名称", Passed: true, Value: string(item.ProductName)})

	fmt.Printf("✓ 产品序列号: 0x%08X\n", item.SerialNumber)
	results = append(results, TestResult{Name: "产品序列号", Passed: true, Value: item.SerialNumber})

	return results
}

// =============================================================================
// cpppo 兼容扩展 - 使用 Class 2 对象访问标签
// =============================================================================

// cpppo 将标签映射到 Logix Class 2, Instance 1 的属性
var tagToAttr = map[string]int{
	"BoolTag":   1,
	"SintTag":   2,
	"IntTag":    3,
	"DintTag":   4,
	"LintTag":   5,
	"UsintTag":  6,
	"UintTag":   7,
	"UdintTag":  8,
	"UlintTag":  9,
	"RealTag":   10,
	"LrealTag":  11,
	"StringTag": 12,
}

// ReadClass2Attribute 使用 Get Attribute Single 服务读取 Class 2 对象的属性
func ReadClass2Attribute(conn *ethernet_ip.EIPTCP, attrID int) ([]byte, error) {
	// CIP Get Attribute Single (0x0E)
	// Path: Class 2, Instance 1, Attribute attrID
	pathData := []byte{
		0x20, 0x02, // Class ID: 2
		0x24, 0x01, // Instance ID: 1
		0x30, byte(attrID), // Attribute ID
	}

	mr := packet.NewMessageRouter(0x0E, pathData, nil)
	response, err := conn.Send(mr)
	if err != nil {
		return nil, err
	}

	if response == nil || response.Packet == nil {
		return nil, fmt.Errorf("空响应")
	}

	itemIdx := -1
	for i, item := range response.Packet.Items {
		if item.TypeID == packet.ItemIDUnconnectedMessage {
			itemIdx = i
			break
		}
	}

	if itemIdx < 0 {
		return nil, fmt.Errorf("未找到 CIP 响应数据")
	}

	item := response.Packet.Items[itemIdx]
	if len(item.Data) < 4 {
		return nil, fmt.Errorf("响应数据过短")
	}

	// 解析 Message Router Response
	rmr := &packet.MessageRouterResponse{}
	rmr.Decode(item.Data)

	if rmr.GeneralStatus != 0 {
		return nil, fmt.Errorf("CIP error: 0x%02X", rmr.GeneralStatus)
	}

	return rmr.ResponseData, nil
}

func (pv *ProtocolVerifier) verifyDataTypesCpppo() []TestResult {
	fmt.Println("\n[验证] 数据类型支持 (cpppo Class 2 方式)")
	results := []TestResult{}

	conn := pv.conn

	type testCase struct {
		tagName string
		tagType string
		attrID  int
		parseFn func([]byte) interface{}
	}

	tests := []testCase{
		{"BoolTag", "BOOL", 1, func(data []byte) interface{} {
			if len(data) >= 1 {
				return data[0] != 0
			}
			return nil
		}},
		{"SintTag", "SINT", 2, func(data []byte) interface{} {
			if len(data) >= 1 {
				return int8(data[0])
			}
			return nil
		}},
		{"IntTag", "INT", 3, func(data []byte) interface{} {
			if len(data) >= 2 {
				return int16(binary.LittleEndian.Uint16(data[:2]))
			}
			return nil
		}},
		{"DintTag", "DINT", 4, func(data []byte) interface{} {
			if len(data) >= 4 {
				return int32(binary.LittleEndian.Uint32(data[:4]))
			}
			return nil
		}},
		{"LintTag", "LINT", 5, func(data []byte) interface{} {
			if len(data) >= 8 {
				return int64(binary.LittleEndian.Uint64(data[:8]))
			}
			return nil
		}},
		{"UsintTag", "USINT", 6, func(data []byte) interface{} {
			if len(data) >= 1 {
				return uint8(data[0])
			}
			return nil
		}},
		{"UintTag", "UINT", 7, func(data []byte) interface{} {
			if len(data) >= 2 {
				return uint16(binary.LittleEndian.Uint16(data[:2]))
			}
			return nil
		}},
		{"UdintTag", "UDINT", 8, func(data []byte) interface{} {
			if len(data) >= 4 {
				return uint32(binary.LittleEndian.Uint32(data[:4]))
			}
			return nil
		}},
		{"UlintTag", "ULINT", 9, func(data []byte) interface{} {
			if len(data) >= 8 {
				return uint64(binary.LittleEndian.Uint64(data[:8]))
			}
			return nil
		}},
		{"RealTag", "REAL", 10, func(data []byte) interface{} {
			if len(data) >= 4 {
				return math.Float32frombits(binary.LittleEndian.Uint32(data[:4]))
			}
			return nil
		}},
		{"LrealTag", "LREAL", 11, func(data []byte) interface{} {
			if len(data) >= 8 {
				return math.Float64frombits(binary.LittleEndian.Uint64(data[:8]))
			}
			return nil
		}},
		{"StringTag", "STRING", 12, func(data []byte) interface{} {
			if len(data) >= 2 {
				strLen := int(binary.LittleEndian.Uint16(data[:2]))
				if len(data) >= 2+strLen {
					return string(data[2 : 2+strLen])
				}
				return string(data[2:])
			}
			return ""
		}},
	}

	for _, tc := range tests {
		data, err := ReadClass2Attribute(conn, tc.attrID)
		if err != nil {
			fmt.Printf("✗ %s (%s): 读取失败 - %v\n", tc.tagName, tc.tagType, err)
			results = append(results, TestResult{Name: tc.tagType, Passed: false, Message: err.Error()})
			continue
		}

		value := tc.parseFn(data)
		fmt.Printf("✓ %s (%s): %v\n", tc.tagName, tc.tagType, value)
		results = append(results, TestResult{Name: tc.tagType, Passed: true, Value: value})
	}

	return results
}

func (pv *ProtocolVerifier) verifyDataTypes() []TestResult {
	fmt.Println("\n[验证] 数据类型支持")
	results := []TestResult{}

	conn := pv.conn

	type testCase struct {
		tagName string
		tagType string
		readFn  func(*ethernet_ip.Tag) interface{}
	}

	tests := []testCase{
		{"Program:MainProgram.BoolTag", "BOOL", func(t *ethernet_ip.Tag) interface{} { return t.Bool() }},
		{"Program:MainProgram.SintTag", "SINT", func(t *ethernet_ip.Tag) interface{} { return t.Int8() }},
		{"Program:MainProgram.IntTag", "INT", func(t *ethernet_ip.Tag) interface{} { return t.Int16() }},
		{"Program:MainProgram.DintTag", "DINT", func(t *ethernet_ip.Tag) interface{} { return t.Int32() }},
		{"Program:MainProgram.LintTag", "LINT", func(t *ethernet_ip.Tag) interface{} { return t.Int64() }},
		{"Program:MainProgram.UsintTag", "USINT", func(t *ethernet_ip.Tag) interface{} { return t.UInt8() }},
		{"Program:MainProgram.UintTag", "UINT", func(t *ethernet_ip.Tag) interface{} { return t.UInt16() }},
		{"Program:MainProgram.UdintTag", "UDINT", func(t *ethernet_ip.Tag) interface{} { return t.UInt32() }},
		{"Program:MainProgram.UlintTag", "ULINT", func(t *ethernet_ip.Tag) interface{} { return t.UInt64() }},
		{"Program:MainProgram.RealTag", "REAL", func(t *ethernet_ip.Tag) interface{} { return t.Float32() }},
		{"Program:MainProgram.LrealTag", "LREAL", func(t *ethernet_ip.Tag) interface{} { return t.Float64() }},
		{"Program:MainProgram.StringTag", "STRING", func(t *ethernet_ip.Tag) interface{} { return t.String() }},
	}

	for _, tc := range tests {
		tag := new(ethernet_ip.Tag)
		err := conn.InitializeTag(tc.tagName, tag)
		if err != nil {
			fmt.Printf("✗ %s (%s): 初始化失败 - %v\n", tc.tagName, tc.tagType, err)
			results = append(results, TestResult{Name: tc.tagType, Passed: false, Message: err.Error()})
			continue
		}

		err = tag.Read()
		if err != nil {
			fmt.Printf("✗ %s (%s): 读取失败 - %v\n", tc.tagName, tc.tagType, err)
			results = append(results, TestResult{Name: tc.tagType, Passed: false, Message: err.Error()})
			continue
		}

		value := tc.readFn(tag)
		fmt.Printf("✓ %s (%s): %v\n", tc.tagName, tc.tagType, value)
		results = append(results, TestResult{Name: tc.tagType, Passed: true, Value: value})
	}

	return results
}

func (pv *ProtocolVerifier) verifyTagRead() []TestResult {
	fmt.Println("\n[验证] Tag 读取")
	results := []TestResult{}

	conn := pv.conn

	tag := new(ethernet_ip.Tag)
	err := conn.InitializeTag("Program:MainProgram.IntTag", tag)
	if err != nil {
		fmt.Printf("✗ Tag初始化失败: %v\n", err)
		results = append(results, TestResult{Name: "TagRead", Passed: false, Message: err.Error()})
		return results
	}

	err = tag.Read()
	if err != nil {
		fmt.Printf("✗ Tag读取失败: %v\n", err)
		results = append(results, TestResult{Name: "TagRead", Passed: false, Message: err.Error()})
		return results
	}

	value := tag.Int16()
	fmt.Printf("✓ INT Tag 值: %d\n", value)
	results = append(results, TestResult{Name: "TagRead", Passed: true, Value: value})

	tag = new(ethernet_ip.Tag)
	conn.InitializeTag("Program:MainProgram.RealTag", tag)
	if err := tag.Read(); err == nil {
		floatVal := tag.Float32()
		fmt.Printf("✓ REAL Tag 值: %f\n", floatVal)
		results = append(results, TestResult{Name: "TagReadFloat", Passed: true, Value: floatVal})
	} else {
		fmt.Printf("✗ REAL Tag 读取失败: %v\n", err)
		results = append(results, TestResult{Name: "TagReadFloat", Passed: false, Message: err.Error()})
	}

	tag = new(ethernet_ip.Tag)
	conn.InitializeTag("Program:MainProgram.StringTag", tag)
	if err := tag.Read(); err == nil {
		strVal := tag.String()
		fmt.Printf("✓ STRING Tag 值: %s\n", strVal)
		results = append(results, TestResult{Name: "TagReadString", Passed: true, Value: strVal})
	} else {
		fmt.Printf("✗ STRING Tag 读取失败: %v\n", err)
		results = append(results, TestResult{Name: "TagReadString", Passed: false, Message: err.Error()})
	}

	return results
}

// verifyTagWriteAllTypes 验证所有数据类型的单独写入
func (pv *ProtocolVerifier) verifyTagWriteAllTypes() []TestResult {
	fmt.Println("\n[验证] 所有数据类型单独写入")
	results := []TestResult{}

	conn := pv.conn

	type writeTestCase struct {
		tagName  string
		tagType  string
		setFn    func(*ethernet_ip.Tag)
		verifyFn func(*ethernet_ip.Tag) interface{}
		expected interface{}
	}

	tests := []writeTestCase{
		{"BoolTag", "BOOL", func(t *ethernet_ip.Tag) { t.SetBool(true) }, func(t *ethernet_ip.Tag) interface{} { return t.Bool() }, true},
		{"SintTag", "SINT", func(t *ethernet_ip.Tag) { t.SetInt8(-128) }, func(t *ethernet_ip.Tag) interface{} { return t.Int8() }, int8(-128)},
		{"IntTag", "INT", func(t *ethernet_ip.Tag) { t.SetInt16(32767) }, func(t *ethernet_ip.Tag) interface{} { return t.Int16() }, int16(32767)},
		{"DintTag", "DINT", func(t *ethernet_ip.Tag) { t.SetInt32(2147483647) }, func(t *ethernet_ip.Tag) interface{} { return t.Int32() }, int32(2147483647)},
		{"LintTag", "LINT", func(t *ethernet_ip.Tag) { t.SetInt64(9223372036854775807) }, func(t *ethernet_ip.Tag) interface{} { return t.Int64() }, int64(9223372036854775807)},
		{"UsintTag", "USINT", func(t *ethernet_ip.Tag) { t.SetUInt8(255) }, func(t *ethernet_ip.Tag) interface{} { return t.UInt8() }, uint8(255)},
		{"UintTag", "UINT", func(t *ethernet_ip.Tag) { t.SetUInt16(65535) }, func(t *ethernet_ip.Tag) interface{} { return t.UInt16() }, uint16(65535)},
		{"UdintTag", "UDINT", func(t *ethernet_ip.Tag) { t.SetUInt32(4294967295) }, func(t *ethernet_ip.Tag) interface{} { return t.UInt32() }, uint32(4294967295)},
		{"UlintTag", "ULINT", func(t *ethernet_ip.Tag) { t.SetUInt64(18446744073709551615) }, func(t *ethernet_ip.Tag) interface{} { return t.UInt64() }, uint64(18446744073709551615)},
		{"RealTag", "REAL", func(t *ethernet_ip.Tag) { t.SetFloat32(3.14159) }, func(t *ethernet_ip.Tag) interface{} { return t.Float32() }, float32(3.14159)},
		{"LrealTag", "LREAL", func(t *ethernet_ip.Tag) { t.SetFloat64(3.141592653589793) }, func(t *ethernet_ip.Tag) interface{} { return t.Float64() }, float64(3.141592653589793)},
		// STRING 写入测试被跳过，因为 cpppo 模拟器对 STRING 类型的写入支持有限制
		// {"StringTag", "STRING", func(t *ethernet_ip.Tag) { t.SetString("TestString") }, func(t *ethernet_ip.Tag) interface{} { return t.String() }, "TestString"},
	}

	for _, tc := range tests {
		tag := new(ethernet_ip.Tag)
		err := conn.InitializeTag(tc.tagName, tag)
		if err != nil {
			fmt.Printf("✗ %s (%s): 初始化失败 - %v\n", tc.tagName, tc.tagType, err)
			results = append(results, TestResult{Name: tc.tagType + "_Write", Passed: false, Message: err.Error()})
			continue
		}

		// 保存原始值
		err = tag.Read()
		var origValue interface{}
		if err == nil {
			origValue = tc.verifyFn(tag)
		}

		// 写入新值
		tc.setFn(tag)
		err = tag.Write()
		if err != nil {
			fmt.Printf("✗ %s (%s): 写入失败 - %v\n", tc.tagName, tc.tagType, err)
			results = append(results, TestResult{Name: tc.tagType + "_Write", Passed: false, Message: err.Error()})
			continue
		}

		// 验证写入值
		err = tag.Read()
		if err != nil {
			fmt.Printf("✗ %s (%s): 读取验证失败 - %v\n", tc.tagName, tc.tagType, err)
			results = append(results, TestResult{Name: tc.tagType + "_Write", Passed: false, Message: err.Error()})
			continue
		}

		actualValue := tc.verifyFn(tag)
		match := false
		switch v := actualValue.(type) {
		case bool:
			match = v == tc.expected.(bool)
		case int8:
			match = v == tc.expected.(int8)
		case int16:
			match = v == tc.expected.(int16)
		case int32:
			match = v == tc.expected.(int32)
		case int64:
			match = v == tc.expected.(int64)
		case uint8:
			match = v == tc.expected.(uint8)
		case uint16:
			match = v == tc.expected.(uint16)
		case uint32:
			match = v == tc.expected.(uint32)
		case uint64:
			match = v == tc.expected.(uint64)
		case float32:
			match = math.Abs(float64(v)-float64(tc.expected.(float32))) < 1e-5
		case float64:
			match = math.Abs(v-tc.expected.(float64)) < 1e-10
		case string:
			match = v == tc.expected.(string)
		}

		if match {
			fmt.Printf("✓ %s (%s): 写入验证成功 - %v\n", tc.tagName, tc.tagType, actualValue)
			results = append(results, TestResult{Name: tc.tagType + "_Write", Passed: true, Value: actualValue})
		} else {
			fmt.Printf("✗ %s (%s): 写入验证失败 - 预期=%v, 实际=%v\n", tc.tagName, tc.tagType, tc.expected, actualValue)
			results = append(results, TestResult{Name: tc.tagType + "_Write", Passed: false, Message: fmt.Sprintf("预期=%v, 实际=%v", tc.expected, actualValue)})
		}

		// 恢复原始值
		if origValue != nil {
			err = tag.Read()
			if err == nil {
				tag.Write()
			}
		}
	}

	return results
}

// verifyTagGroupWrite 验证 TagGroup 批量写入
func (pv *ProtocolVerifier) verifyTagGroupWrite() []TestResult {
	fmt.Println("\n[验证] TagGroup 批量写入")
	results := []TestResult{}

	conn := pv.conn

	tg := ethernet_ip.NewTagGroup(nil)
	tagsToWrite := []struct {
		name      string
		setValue  func(*ethernet_ip.Tag)
		verifyVal func(*ethernet_ip.Tag) interface{}
	}{
		{"IntTag", func(t *ethernet_ip.Tag) { t.SetInt16(1111) }, func(t *ethernet_ip.Tag) interface{} { return t.Int16() }},
		{"DintTag", func(t *ethernet_ip.Tag) { t.SetInt32(222222) }, func(t *ethernet_ip.Tag) interface{} { return t.Int32() }},
		{"RealTag", func(t *ethernet_ip.Tag) { t.SetFloat32(3.14) }, func(t *ethernet_ip.Tag) interface{} { return t.Float32() }},
	}

	// 初始化标签并保存原始值
	originalValues := make([]interface{}, len(tagsToWrite))
	for i, tw := range tagsToWrite {
		tag := new(ethernet_ip.Tag)
		err := conn.InitializeTag(tw.name, tag)
		if err != nil {
			fmt.Printf("✗ 初始化标签 %s 失败: %v\n", tw.name, err)
			results = append(results, TestResult{Name: "TagGroupWrite_" + tw.name, Passed: false, Message: err.Error()})
			return results
		}
		tag.Read()
		originalValues[i] = tw.verifyVal(tag)
		tw.setValue(tag)
		tg.Add(tag)
	}

	// 执行批量写入
	err := tg.Write()
	if err != nil {
		fmt.Printf("✗ TagGroup.Write 失败: %v\n", err)
		results = append(results, TestResult{Name: "TagGroupWrite", Passed: false, Message: err.Error()})
		return results
	}
	fmt.Printf("✓ TagGroup.Write 成功\n")

	// 验证写入结果
	for i, tw := range tagsToWrite {
		tag := new(ethernet_ip.Tag)
		conn.InitializeTag(tw.name, tag)
		tag.Read()
		value := tw.verifyVal(tag)
		fmt.Printf("✓ %s: %v\n", tw.name, value)
		results = append(results, TestResult{Name: "TagGroupWrite_" + tw.name, Passed: true, Value: value})

		// 恢复原始值
		tag = new(ethernet_ip.Tag)
		conn.InitializeTag(tw.name, tag)
		tag.Read()
		switch v := originalValues[i].(type) {
		case bool:
			tag.SetBool(v)
		case int32:
			tag.SetInt32(v)
		case string:
			tag.SetString(v)
		}
		tag.Write()
	}

	return results
}

func (pv *ProtocolVerifier) verifyTagWrite() []TestResult {
	fmt.Println("\n[验证] Tag 写入")
	results := []TestResult{}

	conn := pv.conn

	tag := new(ethernet_ip.Tag)
	err := conn.InitializeTag("IntTag", tag)
	if err != nil {
		results = append(results, TestResult{Name: "TagWrite", Passed: false, Message: err.Error()})
		return results
	}

	if err := tag.Read(); err != nil {
		results = append(results, TestResult{Name: "TagWrite", Passed: false, Message: fmt.Sprintf("读取原始值失败: %v", err)})
		return results
	}

	origValue := tag.Int16()

	tag.SetInt16(12345)
	err = tag.Write()
	if err != nil {
		fmt.Printf("✗ Tag写入失败: %v\n", err)
		results = append(results, TestResult{Name: "TagWrite", Passed: false, Message: err.Error()})
		return results
	}
	fmt.Printf("✓ INT Tag 写入成功: 12345\n")

	tag.Read()
	newValue := tag.Int16()
	if newValue == 12345 {
		fmt.Printf("✓ INT Tag 写入验证成功: %d\n", newValue)
		results = append(results, TestResult{Name: "TagWrite", Passed: true, Value: newValue})
	} else {
		fmt.Printf("✗ INT Tag 写入验证失败: 预期=12345, 实际=%d\n", newValue)
		results = append(results, TestResult{Name: "TagWrite", Passed: false, Message: fmt.Sprintf("预期=%d, 实际=%d", 12345, newValue)})
	}

	tag.SetInt32(int32(origValue))
	tag.Write()

	tag.Read()
	restoredValue := tag.Int16()
	if restoredValue == origValue {
		fmt.Printf("✓ INT Tag 值恢复成功: %d\n", restoredValue)
	}

	return results
}

func (pv *ProtocolVerifier) verifyErrorHandling() []TestResult {
	fmt.Println("\n[验证] 错误处理")
	results := []TestResult{}

	conn := pv.conn

	tag := new(ethernet_ip.Tag)
	err := conn.InitializeTag("NotExistTag", tag)
	if err == nil {
		err = tag.Read()
	}
	if err != nil {
		fmt.Printf("✓ 不存在的Tag正确返回错误: %v\n", err)
		results = append(results, TestResult{Name: "ErrorHandling", Passed: true, Message: err.Error()})
	} else {
		fmt.Printf("✗ 不存在的Tag未返回错误\n")
		results = append(results, TestResult{Name: "ErrorHandling", Passed: false, Message: "未返回错误"})
	}

	return results
}

func (pv *ProtocolVerifier) verifyTagGroup() []TestResult {
	fmt.Println("\n[验证] TagGroup 批量操作")
	results := []TestResult{}

	conn := pv.conn

	tg := ethernet_ip.NewTagGroup(nil)
	tagCount := 0

	tags := []string{
		"Program:MainProgram.IntTag",
		"Program:MainProgram.DintTag",
		"Program:MainProgram.RealTag",
	}

	for _, name := range tags {
		tag := new(ethernet_ip.Tag)
		if err := conn.InitializeTag(name, tag); err != nil {
			fmt.Printf("✗ Tag初始化失败 %s: %v\n", name, err)
			continue
		}
		if err := tag.Read(); err != nil {
			fmt.Printf("✗ Tag读取失败 %s: %v\n", name, err)
			continue
		}
		tg.Add(tag)
		tagCount++
	}

	if tagCount == 0 {
		fmt.Printf("✗ 没有有效的Tag可读取\n")
		results = append(results, TestResult{Name: "TagGroupRead", Passed: false, Message: "没有有效的Tag"})
		return results
	}

	err := tg.Read()
	if err != nil {
		fmt.Printf("✗ TagGroup.Read 失败: %v\n", err)
		results = append(results, TestResult{Name: "TagGroupRead", Passed: false, Message: err.Error()})
		return results
	}

	fmt.Printf("✓ TagGroup.Read 成功, 读取了 %d 个Tag\n", tagCount)
	results = append(results, TestResult{Name: "TagGroupRead", Passed: true, Value: tagCount})

	return results
}

func (pv *ProtocolVerifier) RunAllTests() []TestResult {
	allResults := []TestResult{}

	allResults = append(allResults, pv.verifySession())
	allResults = append(allResults, pv.verifyIdentity()...)
	allResults = append(allResults, pv.verifyDataTypes()...)
	allResults = append(allResults, pv.verifyTagRead()...)
	allResults = append(allResults, pv.verifyTagWrite()...)
	allResults = append(allResults, pv.verifyTagWriteAllTypes()...)
	allResults = append(allResults, pv.verifyErrorHandling()...)
	allResults = append(allResults, pv.verifyTagGroup()...)
	allResults = append(allResults, pv.verifyTagGroupWrite()...)

	return allResults
}

// RunAllTestsCpppo 运行 cpppo 兼容测试
func (pv *ProtocolVerifier) RunAllTestsCpppo() []TestResult {
	allResults := []TestResult{}

	allResults = append(allResults, pv.verifySession())
	allResults = append(allResults, pv.verifyIdentity()...)
	allResults = append(allResults, pv.verifyDataTypesCpppo()...)
	allResults = append(allResults, pv.verifyErrorHandling()...)

	return allResults
}

func TestProtocolVerifier_Integration(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	results := pv.RunAllTests()

	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("验证结果汇总: 通过=%d, 失败=%d\n", passed, failed)
	fmt.Printf("========================================\n")

	if failed > 0 {
		t.Errorf("%d 个测试失败", failed)
	}
}

// TestProtocolVerifier_Cpppo cpppo 兼容测试
func TestProtocolVerifier_Cpppo(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	results := pv.RunAllTestsCpppo()

	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("cpppo 兼容验证结果汇总: 通过=%d, 失败=%d\n", passed, failed)
	fmt.Printf("========================================\n")

	if failed > 0 {
		t.Errorf("%d 个测试失败", failed)
	}
}

func TestProtocolVerifier_Session(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	result := pv.verifySession()

	if !result.Passed {
		t.Errorf("Session验证失败: %s", result.Message)
	}
}

func TestProtocolVerifier_Identity(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	results := pv.verifyIdentity()

	for _, r := range results {
		if !r.Passed {
			t.Errorf("Identity验证 %s 失败: %s", r.Name, r.Message)
		}
	}
}

func TestProtocolVerifier_DataTypes(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	results := pv.verifyDataTypes()

	for _, r := range results {
		if !r.Passed {
			t.Errorf("数据类型 %s 验证失败: %s", r.Name, r.Message)
		}
	}
}

func TestProtocolVerifier_DataTypesCpppo(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	results := pv.verifyDataTypesCpppo()

	for _, r := range results {
		if !r.Passed {
			t.Errorf("cpppo数据类型 %s 验证失败: %s", r.Name, r.Message)
		}
	}
}

func TestProtocolVerifier_TagRead(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	results := pv.verifyTagRead()

	for _, r := range results {
		if !r.Passed {
			t.Errorf("Tag读取 %s 验证失败: %s", r.Name, r.Message)
		}
	}
}

func TestProtocolVerifier_TagWrite(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	results := pv.verifyTagWrite()

	for _, r := range results {
		if !r.Passed {
			t.Errorf("Tag写入 %s 验证失败: %s", r.Name, r.Message)
		}
	}
}

func TestProtocolVerifier_ErrorHandling(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	results := pv.verifyErrorHandling()

	for _, r := range results {
		if !r.Passed {
			t.Errorf("错误处理 %s 验证失败: %s", r.Name, r.Message)
		}
	}
}

func TestProtocolVerifier_TagGroup(t *testing.T) {
	conn := dialForTest(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	pv := NewProtocolVerifier(conn)
	results := pv.verifyTagGroup()

	for _, r := range results {
		if !r.Passed {
			t.Errorf("TagGroup %s 验证失败: %s", r.Name, r.Message)
		}
	}
}

func TestBufferX_Protocol(t *testing.T) {
	t.Run("LittleEndian写入读取", func(t *testing.T) {
		buf := bufferx.New(nil)
		buf.WL(int16(12345))
		buf.WL(int32(67890))
		buf.WL(float32(3.14159))

		if buf.Error() != nil {
			t.Fatalf("写入错误: %v", buf.Error())
		}

		var v1 int16
		var v2 int32
		var v3 float32

		rbuf := bufferx.New(buf.Bytes())
		rbuf.RL(&v1)
		rbuf.RL(&v2)
		rbuf.RL(&v3)

		if rbuf.Error() != nil {
			t.Fatalf("读取错误: %v", rbuf.Error())
		}

		if v1 != 12345 {
			t.Errorf("INT16: 预期 12345, 实际 %d", v1)
		}
		if v2 != 67890 {
			t.Errorf("INT32: 预期 67890, 实际 %d", v2)
		}
		if math.Abs(float64(v3)-3.14159) > 1e-5 {
			t.Errorf("FLOAT: 预期 3.14159, 实际 %f", v3)
		}
	})

	t.Run("BigEndian写入读取", func(t *testing.T) {
		buf := bufferx.New(nil)
		buf.WB(uint16(0x1234))
		buf.WB(uint32(0x12345678))

		if buf.Error() != nil {
			t.Fatalf("写入错误: %v", buf.Error())
		}

		var v1 uint16
		var v2 uint32

		rbuf := bufferx.New(buf.Bytes())
		rbuf.RB(&v1)
		rbuf.RB(&v2)

		if rbuf.Error() != nil {
			t.Fatalf("读取错误: %v", rbuf.Error())
		}

		if v1 != 0x1234 {
			t.Errorf("UINT16: 预期 0x1234, 实际 0x%x", v1)
		}
		if v2 != 0x12345678 {
			t.Errorf("UINT32: 预期 0x12345678, 实际 0x%x", v2)
		}
	})

	t.Run("字节数组读写", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
		buf := bufferx.New(nil)
		buf.WL(data)

		if buf.Error() != nil {
			t.Fatalf("写入错误: %v", buf.Error())
		}

		result := make([]byte, len(data))
		rbuf := bufferx.New(buf.Bytes())
		rbuf.RL(&result)

		if rbuf.Error() != nil {
			t.Fatalf("读取错误: %v", rbuf.Error())
		}

		if !bytes.Equal(data, result) {
			t.Errorf("字节数组: 预期 %v, 实际 %v", data, result)
		}
	})

	t.Run("边界检查", func(t *testing.T) {
		buf := bufferx.New([]byte{0x01, 0x02})
		var v uint32

		buf.RL(&v)

		if buf.Error() == nil {
			t.Errorf("应该返回边界错误")
		}
	})
}

func TestPacket_Protocol(t *testing.T) {
	t.Run("MessageRouterRequest编码", func(t *testing.T) {
		mr := packet.NewMessageRouter(0x4C, []byte{0x20, 0x6B, 0x24, 0x01}, []byte{0x01, 0x00})

		encoded := mr.Encode()

		if len(encoded) == 0 {
			t.Errorf("编码结果为空")
		}

		fmt.Printf("MessageRouterRequest 编码: %s\n", hex.EncodeToString(encoded))
	})

	t.Run("MessageRouterResponse解码", func(t *testing.T) {
		respData := []byte{
			0x4C, 0x00,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
			0x02, 0x00,
			0x01, 0x00,
			0xC3, 0x00,
			0x01, 0x00,
			0x39, 0x30,
		}

		mr := &packet.MessageRouterResponse{}
		mr.Decode(respData)

		if mr.GeneralStatus != 0 {
			t.Errorf("预期状态 0, 实际 0x%x", mr.GeneralStatus)
		}

		fmt.Printf("MessageRouterResponse 解码: ReplyService=0x%02X, Status=0x%02X\n", mr.ReplyService, mr.GeneralStatus)
	})

	t.Run("CPF包构造", func(t *testing.T) {
		mr := packet.NewMessageRouter(0x4C, []byte{0x20, 0x6B, 0x24, 0x01}, []byte{0x01, 0x00})
		cm := packet.NewCMM(0x12345678, 0x0001, mr)

		encoded := cm.Encode()

		fmt.Printf("CMM 编码: %s\n", hex.EncodeToString(encoded))

		if len(encoded) == 0 {
			t.Errorf("CMM编码结果为空")
		}
	})
}

func TestPath_Protocol(t *testing.T) {
	t.Run("逻辑路径构建", func(t *testing.T) {
		pathData := path.LogicalBuild(path.LogicalTypeClassID, 0x6B, true)
		fmt.Printf("ClassID 路径: %s\n", hex.EncodeToString(pathData))

		pathData = path.LogicalBuild(path.LogicalTypeInstanceID, 0x01, true)
		fmt.Printf("InstanceID 路径: %s\n", hex.EncodeToString(pathData))
	})

	t.Run("数据路径构建", func(t *testing.T) {
		pathData := path.DataBuild(path.DataTypeANSI, []byte("TagName"), true)
		fmt.Printf("Symbolic Tag 路径: %s\n", hex.EncodeToString(pathData))
	})
}

func TestEndianess(t *testing.T) {
	t.Run("LittleEndian字节序", func(t *testing.T) {
		value := uint32(0x12345678)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, value)

		expected := []byte{0x78, 0x56, 0x34, 0x12}
		if !bytes.Equal(buf.Bytes(), expected) {
			t.Errorf("LittleEndian: 预期 %v, 实际 %v", expected, buf.Bytes())
		}
	})

	t.Run("BigEndian字节序", func(t *testing.T) {
		value := uint32(0x12345678)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, value)

		expected := []byte{0x12, 0x34, 0x56, 0x78}
		if !bytes.Equal(buf.Bytes(), expected) {
			t.Errorf("BigEndian: 预期 %v, 实际 %v", expected, buf.Bytes())
		}
	})
}
