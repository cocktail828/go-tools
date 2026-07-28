package spinner

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// safeBuffer 是并发安全的 bytes.Buffer，供 spinner goroutine 与测试主协程共享。
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// fastOpts 用极小的时间参数加速测试，避免逐字展示拖慢用例。
func fastOpts(buf *safeBuffer, extra ...Option) []Option {
	opts := []Option{
		WithWriter(buf),
		WithCharInterval(time.Millisecond),
		WithFrameRate(2 * time.Millisecond),
		WithHoldDuration(5 * time.Millisecond),
	}
	return append(opts, extra...)
}

func TestSpinnerStatic(t *testing.T) {
	buf := &safeBuffer{}
	msgs := []string{"正在加载模型", "正在执行推理"}

	s := New(msgs, fastOpts(buf)...)
	s.Start()
	// 足够轮播完所有消息若干轮。
	time.Sleep(300 * time.Millisecond)
	s.Stop()

	out := buf.String()
	for _, m := range msgs {
		if !strings.Contains(out, m) {
			t.Errorf("输出中未包含消息 %q", m)
		}
	}
}

func TestSpinnerDynamic(t *testing.T) {
	ch := make(chan string)

	// 动态模式要求至少一条初始文本。
	s := New([]string{"准备中"}, WithUpdates(ch))
	s.Start()

	send := func(text string) {
		ch <- text
	}

	send("正在加载模型")
	time.Sleep(time.Millisecond * 500)
	send("正在执行推理")
	time.Sleep(time.Millisecond * 500)
	send("正在保存结果alsfhlasdflasjdflasdlfklaksjdf")
	time.Sleep(time.Millisecond * 5000)

	s.Stop()
}

func TestSpinnerDynamicInitialMessage(t *testing.T) {
	buf := &safeBuffer{}
	ch := make(chan string)

	// 动态模式下 messages[0] 作为初始占位文本。
	s := New([]string{"准备中"}, fastOpts(buf, WithUpdates(ch))...)
	s.Start()

	time.Sleep(100 * time.Millisecond)
	if got := buf.String(); !strings.Contains(got, "准备中") {
		t.Errorf("动态模式未展示初始占位文本，实际=%q", got)
	}

	s.Stop()
}

func TestSpinnerDynamicChannelClose(t *testing.T) {
	buf := &safeBuffer{}
	ch := make(chan string)

	// 动态模式要求至少一条初始文本。
	s := New([]string{"准备中"}, fastOpts(buf, WithUpdates(ch))...)
	s.Start()

	ch <- "最后一条"
	time.Sleep(100 * time.Millisecond)
	close(ch)

	// 通道关闭后应继续运行并保留最后文本，Stop 不应阻塞。
	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("通道关闭后 Stop 阻塞了")
	}

	if got := buf.String(); !strings.Contains(got, "最后一条") {
		t.Errorf("通道关闭后未保留最后文本，实际=%q", got)
	}
}

func TestSpinnerStopClearsLine(t *testing.T) {
	buf := &safeBuffer{}
	s := New([]string{"内容"}, fastOpts(buf)...)
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	// Stop 时会写入清行序列 \r\033[K。
	if !strings.HasSuffix(buf.String(), "\r\033[K") {
		t.Errorf("Stop 未以清行序列结尾，实际结尾=%q", tail(buf.String(), 8))
	}
}

func tail(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[len(r)-n:])
}

// TestSpinnerDynamicDemo 是动态模式的可视化演示（非断言）：写到 stdout、正常速度，
// 通过通道模拟外部逐步推进的任务状态。默认在 -short 下跳过。
func TestSpinnerDynamicDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过可视化演示；去掉 -short 可运行")
	}

	ch := make(chan string)
	s := New([]string{"准备中"}, WithUpdates(ch), WithStyle(StyleDots))
	s.Start()

	steps := []string{
		"正在加载模型",
		"正在初始化 KV Cache",
		"正在等待 GPU 资源",
		"正在执行推理",
		"正在保存结果",
	}
	for _, step := range steps {
		ch <- step
		// 停留一会儿观察当前文本；期间不发新文本，spinner 持续转动。
		time.Sleep(2 * time.Second)
	}

	close(ch)
	s.Stop()
	fmt.Println("✅ 完成")
}

// TestSpinnerDemo 是可视化演示（非断言），逐个风格轮播展示。
// 默认在 -short 下跳过，避免拖慢常规测试。
func TestSpinnerDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过可视化演示；去掉 -short 可运行")
	}

	for _, sty := range []Style{
		StyleBreathStar,
		StyleDots,
		StyleCircles,
		StyleStars,
		StyleLine,
		StyleBraille,
		StyleMoon,
		StyleClock,
		StyleGrow,
		StylePulse,
	} {
		s := New([]string{
			"正在加载模型",
			"正在初始化 KV Cache",
			"正在等待 GPU 资源",
			"正在执行推理",
			"正在保存结果",
		}, WithStyle(sty))

		s.Start()
		time.Sleep(2 * time.Second)
		s.Stop()
		fmt.Println("✅ 完成")
	}
}
