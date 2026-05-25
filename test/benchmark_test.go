package test

import (
	"testing"
	"time"

	ethernet_ip "github.com/anviod/ethernet-ip"
)

// BenchmarkTagRead benchmarks reading a tag multiple times
func BenchmarkTagRead(b *testing.B) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		b.Skipf("无法创建连接: %v", err)
		return
	}
	err = conn.Connect()
	if err != nil {
		b.Skipf("无法连接到模拟器: %v", err)
		return
	}
	defer conn.Close()

	tag := new(ethernet_ip.Tag)
	err = conn.InitializeTag("IntTag", tag)
	if err != nil {
		b.Skipf("无法初始化标签: %v", err)
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := tag.Read(); err != nil {
			b.Fatalf("读取失败: %v", err)
		}
	}
}

// BenchmarkTagWrite benchmarks writing a tag multiple times
func BenchmarkTagWrite(b *testing.B) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		b.Skipf("无法创建连接: %v", err)
		return
	}
	err = conn.Connect()
	if err != nil {
		b.Skipf("无法连接到模拟器: %v", err)
		return
	}
	defer conn.Close()

	tag := new(ethernet_ip.Tag)
	err = conn.InitializeTag("IntTag", tag)
	if err != nil {
		b.Skipf("无法初始化标签: %v", err)
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tag.SetInt32(int32(i))
		if err := tag.Write(); err != nil {
			b.Fatalf("写入失败: %v", err)
		}
	}
}

// BenchmarkTagReadWrite benchmarks read-write cycles
func BenchmarkTagReadWrite(b *testing.B) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		b.Skipf("无法创建连接: %v", err)
		return
	}
	err = conn.Connect()
	if err != nil {
		b.Skipf("无法连接到模拟器: %v", err)
		return
	}
	defer conn.Close()

	tag := new(ethernet_ip.Tag)
	err = conn.InitializeTag("IntTag", tag)
	if err != nil {
		b.Skipf("无法初始化标签: %v", err)
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tag.SetInt32(int32(i))
		if err := tag.Write(); err != nil {
			b.Fatalf("写入失败: %v", err)
		}
		if err := tag.Read(); err != nil {
			b.Fatalf("读取失败: %v", err)
		}
	}
}

// BenchmarkTagGroupRead benchmarks reading multiple tags simultaneously
func BenchmarkTagGroupRead(b *testing.B) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		b.Skipf("无法创建连接: %v", err)
		return
	}
	err = conn.Connect()
	if err != nil {
		b.Skipf("无法连接到模拟器: %v", err)
		return
	}
	defer conn.Close()

	tg := ethernet_ip.NewTagGroup(nil)

	tags := []string{"IntTag", "DintTag", "RealTag"}
	for _, name := range tags {
		tag := new(ethernet_ip.Tag)
		if err := conn.InitializeTag(name, tag); err != nil {
			b.Skipf("无法初始化标签 %s: %v", name, err)
			return
		}
		tg.Add(tag)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := tg.Read(); err != nil {
			b.Fatalf("TagGroup读取失败: %v", err)
		}
	}
}

// BenchmarkStringWrite benchmarks string write operations
func BenchmarkStringWrite(b *testing.B) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		b.Skipf("无法创建连接: %v", err)
		return
	}
	err = conn.Connect()
	if err != nil {
		b.Skipf("无法连接到模拟器: %v", err)
		return
	}
	defer conn.Close()

	tag := new(ethernet_ip.Tag)
	err = conn.InitializeTag("StringTag", tag)
	if err != nil {
		b.Skipf("无法初始化标签: %v", err)
		return
	}

	testString := "TestString1234567890"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tag.SetString(testString)
		if err := tag.Write(); err != nil {
			b.Fatalf("字符串写入失败: %v", err)
		}
	}
}

// BenchmarkSessionRegister benchmarks session registration
func BenchmarkSessionRegister(b *testing.B) {
	for i := 0; i < b.N; i++ {
		conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
		if err != nil {
			b.Skipf("无法创建连接: %v", err)
			return
		}
		err = conn.Connect()
		if err != nil {
			b.Skipf("无法连接到模拟器: %v", err)
			return
		}
		conn.Close()
	}
}

// BenchmarkConcurrentTagAccess benchmarks concurrent tag access
func BenchmarkConcurrentTagAccess(b *testing.B) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		b.Skipf("无法创建连接: %v", err)
		return
	}
	err = conn.Connect()
	if err != nil {
		b.Skipf("无法连接到模拟器: %v", err)
		return
	}
	defer conn.Close()

	tag := new(ethernet_ip.Tag)
	err = conn.InitializeTag("IntTag", tag)
	if err != nil {
		b.Skipf("无法初始化标签: %v", err)
		return
	}

	b.SetParallelism(10)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tag.Read()
			time.Sleep(time.Millisecond)
		}
	})
}

// TestPerformanceMetrics collects performance metrics for documentation
func TestPerformanceMetrics(t *testing.T) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		t.Skipf("跳过测试: 无法创建连接: %v", err)
		return
	}
	err = conn.Connect()
	if err != nil {
		t.Skipf("跳过测试: 无法连接到模拟器: %v", err)
		return
	}
	defer conn.Close()

	tag := new(ethernet_ip.Tag)
	err = conn.InitializeTag("IntTag", tag)
	if err != nil {
		t.Skipf("跳过测试: 无法初始化标签: %v", err)
		return
	}

	iterations := 1000

	start := time.Now()
	for i := 0; i < iterations; i++ {
		tag.Read()
	}
	readDuration := time.Since(start)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		tag.SetInt32(int32(i))
		tag.Write()
	}
	writeDuration := time.Since(start)

	t.Logf("=== 性能指标 ===")
	t.Logf("读取 1000 次: %v (平均 %.2f ms/次)", readDuration, float64(readDuration.Milliseconds())/float64(iterations))
	t.Logf("写入 1000 次: %v (平均 %.2f ms/次)", writeDuration, float64(writeDuration.Milliseconds())/float64(iterations))
	t.Logf("总吞吐量: 约 %.2f 次/秒", float64(iterations*2)*1000/float64(readDuration.Milliseconds()+writeDuration.Milliseconds()))
}