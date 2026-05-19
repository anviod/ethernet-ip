package ethernet_ip

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

type performanceResult struct {
	TestName    string  `json:"test_name"`
	Iterations  int     `json:"iterations"`
	TotalTimeMs float64 `json:"total_time_ms"`
	AvgTimeUs   float64 `json:"avg_time_us"`
	Throughput  float64 `json:"throughput_per_sec"`
}

func TestTagWritePerformance(t *testing.T) {
	conn, err := NewTCP("127.0.0.1", nil)
	if err != nil {
		t.Skip("无法创建TCP连接，跳过性能测试")
	}
	if err := conn.Connect(); err != nil {
		t.Skip("无法连接到PLC模拟器，跳过性能测试")
	}
	defer conn.Close()

	var results []performanceResult

	// 单独写入性能测试
	testIterations := []int{10, 50, 100, 500, 1000}
	for _, iterations := range testIterations {
		tag := new(Tag)
		err := conn.InitializeTag("IntTag", tag)
		if err != nil {
			t.Fatalf("初始化标签失败: %v", err)
		}

		start := time.Now()
		for i := 0; i < iterations; i++ {
			tag.SetInt32(int32(i))
			tag.Write()
		}
		duration := time.Since(start)

		avgTimeUs := float64(duration.Microseconds()) / float64(iterations)
		throughput := float64(iterations) / duration.Seconds()

		results = append(results, performanceResult{
			TestName:    fmt.Sprintf("单独写入_%d次", iterations),
			Iterations:  iterations,
			TotalTimeMs: duration.Seconds() * 1000,
			AvgTimeUs:   avgTimeUs,
			Throughput:  throughput,
		})
	}

	// 批量写入性能测试
	tagCounts := []int{10, 50, 100}
	for _, count := range tagCounts {
		tg := NewTagGroup(nil)
		tg.Tcp = conn

		tags := make([]*Tag, count)
		for i := 0; i < count; i++ {
			tags[i] = new(Tag)
			conn.InitializeTag(fmt.Sprintf("Global.IntArray[%d]", i%5), tags[i])
			tg.Add(tags[i])
		}

		start := time.Now()
		for iteration := 0; iteration < 10; iteration++ {
			for i, t := range tags {
				t.SetInt32(int32(iteration*1000 + i))
			}
			tg.Write()
		}
		duration := time.Since(start)

		totalTags := count * 10
		avgTimeUs := float64(duration.Microseconds()) / float64(totalTags)
		throughput := float64(totalTags) / duration.Seconds()

		results = append(results, performanceResult{
			TestName:    fmt.Sprintf("批量写入_%d标签", count),
			Iterations:  totalTags,
			TotalTimeMs: duration.Seconds() * 1000,
			AvgTimeUs:   avgTimeUs,
			Throughput:  throughput,
		})
	}

	// 生成 JSON 报告
	reportData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatalf("生成报告失败: %v", err)
	}

	err = os.WriteFile("doc/performance_report.json", reportData, 0644)
	if err != nil {
		t.Fatalf("写入报告文件失败: %v", err)
	}

	// 生成 Markdown 报告
	generateMarkdownReport(results)
}

func generateMarkdownReport(results []performanceResult) error {
	file, err := os.Create("doc/performance_report.md")
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "# EtherNet/IP 写入性能测试报告")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "## 测试环境")
	fmt.Fprintln(file, "- **日期**:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(file, "- **目标**: 本地 cpppo 模拟器 (127.0.0.1:44818)")
	fmt.Fprintln(file, "")

	fmt.Fprintln(file, "## 单独写入性能测试")
	fmt.Fprintln(file, "| 测试场景 | 迭代次数 | 总耗时(ms) | 平均耗时(μs/次) | 吞吐量(次/秒) |")
	fmt.Fprintln(file, "|---------|---------|-----------|----------------|---------------|")
	for _, r := range results {
		if strings.Contains(r.TestName, "单独") {
			fmt.Fprintf(file, "| %s | %d | %.2f | %.2f | %.2f |\n",
				r.TestName, r.Iterations, r.TotalTimeMs, r.AvgTimeUs, r.Throughput)
		}
	}

	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "## 批量写入性能测试")
	fmt.Fprintln(file, "| 测试场景 | 标签总数 | 总耗时(ms) | 平均耗时(μs/标签) | 吞吐量(标签/秒) |")
	fmt.Fprintln(file, "|---------|---------|-----------|------------------|-----------------|")
	for _, r := range results {
		if strings.Contains(r.TestName, "批量") {
			fmt.Fprintf(file, "| %s | %d | %.2f | %.2f | %.2f |\n",
				r.TestName, r.Iterations, r.TotalTimeMs, r.AvgTimeUs, r.Throughput)
		}
	}

	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "## 测试结论")
	fmt.Fprintln(file, "- 单独写入适用于少量标签的实时更新场景")
	fmt.Fprintln(file, "- 批量写入适用于大量标签的批量同步场景")
	fmt.Fprintln(file, "- 建议在实际应用中根据标签数量选择合适的写入方式")

	return nil
}
