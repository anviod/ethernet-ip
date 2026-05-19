package ethernet_ip

import (
	"sync"
	"testing"

	"github.com/anviod/ethernet-ip/types"
)

// TestAccessMode_CIPMode 测试标准 CIP 模式的标签访问
func TestAccessMode_CIPMode(t *testing.T) {
	testCases := []struct {
		name       string
		tagType    types.UInt
		setupFunc  func(*Tag)
		verifyFunc func(*testing.T, *Tag)
	}{
		{
			name:    "CIP模式读取INT标签",
			tagType: INT,
			setupFunc: func(tag *Tag) {
				tag.Type = INT
				tag.value = []byte{0x39, 0x30} // 12345 in little-endian
			},
			verifyFunc: func(t *testing.T, tag *Tag) {
				result := tag.Int16()
				if result != 12345 {
					t.Errorf("CIP模式INT读取失败: 预期=12345, 实际=%d", result)
				}
			},
		},
		{
			name:    "CIP模式读取DINT标签",
			tagType: DINT,
			setupFunc: func(tag *Tag) {
				tag.Type = DINT
				tag.value = []byte{0x01, 0x02, 0x03, 0x04} // 0x04030201 = 67305985
			},
			verifyFunc: func(t *testing.T, tag *Tag) {
				result := tag.Int32()
				expected := int32(0x04030201)
				if result != expected {
					t.Errorf("CIP模式DINT读取失败: 预期=%d, 实际=%d", expected, result)
				}
			},
		},
		{
			name:    "CIP模式读取BOOL标签",
			tagType: BOOL,
			setupFunc: func(tag *Tag) {
				tag.Type = BOOL
				tag.value = []byte{0x01}
			},
			verifyFunc: func(t *testing.T, tag *Tag) {
				result := tag.Bool()
				if result != true {
					t.Error("CIP模式BOOL读取失败: 预期=true, 实际=false")
				}
			},
		},
		{
			name:    "CIP模式读取STRING标签",
			tagType: STRING,
			setupFunc: func(tag *Tag) {
				tag.Type = STRING
				tag.value = []byte{0x05, 0x00, 0x00, 0x00, 'H', 'e', 'l', 'l', 'o'} // "Hello"
			},
			verifyFunc: func(t *testing.T, tag *Tag) {
				result := tag.String()
				if result != "Hello" {
					t.Errorf("CIP模式STRING读取失败: 预期=Hello, 实际=%s", result)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &Tag{
				Type: tc.tagType,
				Lock: new(sync.Mutex),
			}
			tc.setupFunc(tag)
			tc.verifyFunc(t, tag)
		})
	}
}

// TestAccessMode_CIPModeWrite 测试标准 CIP 模式的标签写入
func TestAccessMode_CIPModeWrite(t *testing.T) {
	testCases := []struct {
		name     string
		tagType  types.UInt
		setValue interface{}
		expected []byte
	}{
		{
			name:     "CIP模式写入DINT",
			tagType:  DINT,
			setValue: int32(12345),
			expected: []byte{0x39, 0x30, 0x00, 0x00}, // 12345 in little-endian
		},
		{
			name:     "CIP模式写入STRING",
			tagType:  STRING,
			setValue: "Test",
			expected: []byte{'T', 'e', 's', 't'},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &Tag{
				Type: tc.tagType,
				Lock: new(sync.Mutex),
			}

			// 设置值
			switch v := tc.setValue.(type) {
			case int32:
				tag.SetInt32(v)
			case string:
				tag.SetString(v)
			}

			// 验证写入值
			if len(tag.wValue) != len(tc.expected) {
				t.Errorf("写入值长度不匹配: 预期=%d, 实际=%d", len(tc.expected), len(tag.wValue))
				return
			}

			for i := range tc.expected {
				if tag.wValue[i] != tc.expected[i] {
					t.Errorf("写入值内容不匹配: 索引=%d, 预期=0x%02X, 实际=0x%02X", i, tc.expected[i], tag.wValue[i])
				}
			}
		})
	}
}

// TestAccessMode_LogixModeAttributeMapping 测试 Logix 模式的属性映射
func TestAccessMode_LogixModeAttributeMapping(t *testing.T) {
	// 标签名到属性 ID 的映射（Logix Class 2 模式）
	expectedMapping := map[string]int{
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

	// 验证 tagToAttr 映射的完整性
	for tagName, expectedAttrID := range expectedMapping {
		actualAttrID, exists := tagToAttr[tagName]
		if !exists {
			t.Errorf("Logix模式属性映射缺失: tagName=%s", tagName)
			continue
		}
		if actualAttrID != expectedAttrID {
			t.Errorf("Logix模式属性映射错误: tagName=%s, 预期attrID=%d, 实际attrID=%d", tagName, expectedAttrID, actualAttrID)
		}
	}
}

// TestAccessMode_ModeComparison 测试两种模式的基本功能
func TestAccessMode_ModeComparison(t *testing.T) {
	// 测试标准 CIP 模式标签初始化
	t.Run("CIP模式标签初始化", func(t *testing.T) {
		tag := &Tag{
			Type: INT,
			Lock: new(sync.Mutex),
		}
		tag.name = []byte("Program:MainProgram.TestTag")

		if tag.Name() != "Program:MainProgram.TestTag" {
			t.Errorf("CIP模式标签名称设置失败: %s", tag.Name())
		}
		if tag.Type != INT {
			t.Error("CIP模式标签类型设置失败")
		}
	})

	// 测试 Logix 模式属性访问
	t.Run("Logix模式属性访问", func(t *testing.T) {
		// 验证所有预期的标签都有映射
		expectedTags := []string{"BoolTag", "SintTag", "IntTag", "DintTag", "LintTag",
			"UsintTag", "UintTag", "UdintTag", "UlintTag", "RealTag", "LrealTag", "StringTag"}
		for _, tagName := range expectedTags {
			if _, exists := tagToAttr[tagName]; !exists {
				t.Errorf("Logix模式缺少属性映射: tagName=%s", tagName)
			}
		}
	})
}

// TestAccessMode_CIPModeTagPathParsing 测试 CIP 模式标签路径解析
func TestAccessMode_CIPModeTagPathParsing(t *testing.T) {
	testCases := []struct {
		name     string
		tagPath  string
		expected string
	}{
		{
			name:     "简单标签路径",
			tagPath:  "MyTag",
			expected: "MyTag",
		},
		{
			name:     "程序范围内标签",
			tagPath:  "Program:MainProgram.Counter",
			expected: "Program:MainProgram.Counter",
		},
		{
			name:     "UDT标签路径",
			tagPath:  "OP.UDT_Alarm.DINT_065_096",
			expected: "OP.UDT_Alarm.DINT_065_096",
		},
		{
			name:     "数组标签路径",
			tagPath:  "Global.IntArray[0]",
			expected: "Global.IntArray[0]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &Tag{
				Lock: new(sync.Mutex),
			}
			tag.name = []byte(tc.tagPath)
			if tag.Name() != tc.expected {
				t.Errorf("标签路径解析失败: 预期=%s, 实际=%s", tc.expected, tag.Name())
			}
		})
	}
}

// TestAccessMode_CIPvsLogixComparison 对比测试两种模式的差异
func TestAccessMode_CIPvsLogixComparison(t *testing.T) {
	t.Run("CIP模式支持UDT", func(t *testing.T) {
		// CIP模式支持UDT标签路径
		tag := &Tag{
			Type: DINT,
			Lock: new(sync.Mutex),
		}
		tag.name = []byte("MyUDT.Field1")
		// 验证标签名称可以包含UDT路径
		if !contains(tag.Name(), "MyUDT") {
			t.Error("CIP模式应该支持UDT标签路径")
		}
	})

	t.Run("Logix模式仅支持预定义属性", func(t *testing.T) {
		// Logix模式只支持预定义的12个属性
		if len(tagToAttr) != 12 {
			t.Errorf("Logix模式应该只支持12个预定义属性: 实际=%d", len(tagToAttr))
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
