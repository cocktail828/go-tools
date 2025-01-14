package metrics_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/metrics"
)

func TestMetrics(t *testing.T) {
	// 创建计数器
	reqCount := metrics.NewCounter()
	successCount := metrics.NewCounter()
	latency := metrics.NewHistogram(metrics.NewExpDecaySample(1028, 0.015)) // 指数衰减样本

	// 注册到默认注册表
	metrics.Register("requests", reqCount)
	metrics.Register("successes", successCount)
	metrics.Register("latency", latency)

	// 模拟请求处理
	for i := 0; i < 1000; i++ {
		reqCount.Inc(1)
		// start := time.Now()

		// 模拟随机成功率
		if i%10 != 0 { // 90% 成功率
			successCount.Inc(1)
		}

		// 模拟随机延迟
		dur := time.Duration(10+5*(i%5)) * time.Millisecond
		time.Sleep(dur)
		latency.Update(int64(dur / time.Millisecond))
	}

	// 打印统计结果
	fmt.Printf("Total requests: %d\n", reqCount.Count())
	fmt.Printf("Success rate: %.2f%%\n", float64(successCount.Count())/float64(reqCount.Count())*100)
	fmt.Printf("QPS: %.2f\n", float64(reqCount.Count())/(float64(1000)*10/1000))
	fmt.Printf("P99 Latency: %.2fms\n", latency.Percentile(0.99))
}
