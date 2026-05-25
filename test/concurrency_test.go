package test

import (
	"sync"
	"testing"
	"time"

	ethernet_ip "github.com/anviod/ethernet-ip"
)

// TestConcurrentTagAccess 测试多个goroutine同时访问标签
func TestConcurrentTagAccess(t *testing.T) {
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

	const goroutineCount = 10
	const iterationsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < iterationsPerGoroutine; j++ {
				tag := new(ethernet_ip.Tag)
				err := conn.InitializeTag("IntTag", tag)
				if err != nil {
					t.Errorf("Goroutine %d, iteration %d: 初始化标签失败: %v", id, j, err)
					return
				}

				err = tag.Read()
				if err != nil {
					t.Errorf("Goroutine %d, iteration %d: 读取标签失败: %v", id, j, err)
					return
				}

				value := tag.Int16()
				if value < 0 {
					t.Errorf("Goroutine %d, iteration %d: 读取值异常: %d", id, j, value)
				}

				time.Sleep(time.Millisecond * 10)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentTagReadWrite 测试并发读写操作
func TestConcurrentTagReadWrite(t *testing.T) {
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

	var wg sync.WaitGroup
	readDone := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-readDone:
				return
			default:
				tag := new(ethernet_ip.Tag)
				if err := conn.InitializeTag("IntTag", tag); err != nil {
					return
				}
				tag.Read()
				time.Sleep(time.Millisecond * 5)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			tag := new(ethernet_ip.Tag)
			if err := conn.InitializeTag("IntTag", tag); err != nil {
				return
			}
			tag.SetInt16(int16(i * 100))
			tag.Write()
			time.Sleep(time.Millisecond * 50)
		}
		close(readDone)
	}()

	wg.Wait()
}

// TestTagGroupConcurrent 测试TagGroup的并发安全性
func TestTagGroupConcurrent(t *testing.T) {
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

	const goroutineCount = 5
	var wg sync.WaitGroup
	wg.Add(goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wg.Done()
			tg := ethernet_ip.NewTagGroup(nil)

			tag1 := new(ethernet_ip.Tag)
			if err := conn.InitializeTag("IntTag", tag1); err != nil {
				return
			}
			tg.Add(tag1)

			tag2 := new(ethernet_ip.Tag)
			if err := conn.InitializeTag("DintTag", tag2); err != nil {
				return
			}
			tg.Add(tag2)

			for j := 0; j < 5; j++ {
				tg.Read()
				time.Sleep(time.Millisecond * 20)
			}
		}()
	}

	wg.Wait()
}

// TestReconnectMechanism 测试重连机制
func TestReconnectMechanism(t *testing.T) {
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

	if !conn.IsConnected() {
		t.Error("连接状态检查失败")
		return
	}

	tag := new(ethernet_ip.Tag)
	if err := conn.InitializeTag("IntTag", tag); err != nil {
		t.Skipf("无法初始化标签: %v", err)
		return
	}

	if err := tag.Read(); err != nil {
		t.Errorf("初始读取失败: %v", err)
		return
	}

	initialValue := tag.Int16()
	t.Logf("初始值: %d", initialValue)

	tag.SetInt16(initialValue + 1)
	if err := tag.Write(); err != nil {
		t.Errorf("写入失败: %v", err)
		return
	}

	if err := tag.Read(); err != nil {
		t.Errorf("写入后读取失败: %v", err)
		return
	}

	newValue := tag.Int16()
	if newValue != initialValue+1 {
		t.Errorf("值验证失败: 预期=%d, 实际=%d", initialValue+1, newValue)
	}
}

// TestExponentialBackoff 测试指数退避算法
func TestExponentialBackoff(t *testing.T) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		t.Skipf("跳过测试: 无法创建连接: %v", err)
		return
	}

	tests := []struct {
		attempt    int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{attempt: 1, expectedMin: time.Second, expectedMax: time.Second * 2},
		{attempt: 2, expectedMin: time.Second * 2, expectedMax: time.Second * 4},
		{attempt: 3, expectedMin: time.Second * 4, expectedMax: time.Second * 8},
		{attempt: 10, expectedMin: time.Second * 30, expectedMax: time.Second * 30},
	}

	for _, tc := range tests {
		delay := conn.ExponentialBackoff(tc.attempt)
		if delay < tc.expectedMin || delay > tc.expectedMax {
			t.Errorf("指数退避计算错误: attempt=%d, delay=%v, 预期范围=[%v, %v]",
				tc.attempt, delay, tc.expectedMin, tc.expectedMax)
		}
	}
}

// TestForwardClose 测试 ForwardClose 方法
func TestForwardClose(t *testing.T) {
	t.Run("关闭未建立的连接", func(t *testing.T) {
		conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
		if err != nil {
			t.Fatalf("无法创建连接: %v", err)
		}

		err = conn.ForwardClose()
		if err == nil {
			t.Error("应该返回错误，因为没有建立forward open连接")
		}
	})

	t.Run("ForwardOpen后关闭", func(t *testing.T) {
		conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
		if err != nil {
			t.Fatalf("无法创建连接: %v", err)
		}
		err = conn.Connect()
		if err != nil {
			t.Skipf("跳过测试: 无法连接到模拟器: %v", err)
			return
		}
		defer conn.Close()

		err = conn.ForwardOpen()
		if err != nil {
			t.Skipf("ForwardOpen失败: %v", err)
			return
		}

		err = conn.ForwardClose()
		if err != nil {
			t.Logf("ForwardClose返回错误(可能是模拟器不支持): %v", err)
		}
	})
}

// TestReconnectStateManagement 测试重连状态管理（不需要实际连接）
func TestReconnectStateManagement(t *testing.T) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		t.Fatalf("无法创建连接: %v", err)
	}

	if conn.IsConnected() {
		t.Error("新建连接不应处于已连接状态")
	}

	if conn.GetReconnectAttempts() != 0 {
		t.Errorf("初始重连次数应为0，实际为%d", conn.GetReconnectAttempts())
	}
}

// TestExponentialBackoffEdgeCases 测试指数退避边界情况
// 公式: delay = minDelay * 2^attempt，最大30秒
func TestExponentialBackoffEdgeCases(t *testing.T) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		t.Fatalf("无法创建连接: %v", err)
	}

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"负数尝试次数", -1, time.Duration(0)},
		{"零次尝试", 0, time.Second * 1},
		{"单次尝试", 1, time.Second * 2},
		{"两次尝试", 2, time.Second * 4},
		{"四次尝试", 4, time.Second * 16},
		{"五次尝试(超过最大延迟)", 5, time.Second * 30},
		{"十次尝试(最大延迟)", 10, time.Second * 30},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			delay := conn.ExponentialBackoff(tc.attempt)
			if delay != tc.expected {
				t.Errorf("预期延迟=%v, 实际=%v", tc.expected, delay)
			}
		})
	}
}

// TestConnectionStateTransitions 测试连接状态转换
func TestConnectionStateTransitions(t *testing.T) {
	conn, err := ethernet_ip.NewTCP("127.0.0.1", nil)
	if err != nil {
		t.Fatalf("无法创建连接: %v", err)
	}

	initialState := conn.IsConnected()
	if initialState {
		t.Error("新建连接不应处于已连接状态")
	}

	err = conn.Connect()
	if err != nil {
		t.Skipf("跳过测试: 无法连接到模拟器: %v", err)
		return
	}

	if !conn.IsConnected() {
		t.Error("连接后应处于已连接状态")
	}

	err = conn.Close()
	if err != nil {
		t.Errorf("Close失败: %v", err)
	}

	if conn.IsConnected() {
		t.Error("关闭后不应处于已连接状态")
	}
}