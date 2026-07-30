package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Option 用于配置 Spinner。
type Option func(*Spinner)

// WithStyle 设置动画展示风格，非法风格（无帧）将被忽略。
func WithStyle(style Style) Option {
	return func(s *Spinner) {
		if style.valid() {
			s.style = style
		}
	}
}

// WithWriter 设置输出目标，默认 os.Stdout。
func WithWriter(w io.Writer) Option {
	return func(s *Spinner) {
		if w != nil {
			s.writer = w
		}
	}
}

// WithCharInterval 设置逐字打字的间隔。
func WithCharInterval(d time.Duration) Option {
	return func(s *Spinner) {
		if d > 0 {
			s.charInterval = d
		}
	}
}

// WithHoldDuration 设置一句消息完整展示后的停留时长。
func WithHoldDuration(d time.Duration) Option {
	return func(s *Spinner) {
		if d > 0 {
			s.holdDuration = d
		}
	}
}

// WithFrameRate 设置动画帧刷新间隔。
func WithFrameRate(d time.Duration) Option {
	return func(s *Spinner) {
		if d > 0 {
			s.frameRate = d
		}
	}
}

// WithUpdates 启用动态模式：文本由外部通道驱动，而非按固定列表轮播。
// 收到新文本时重新逐字展示，无新文本则保持当前文本（spinner 动画持续）。
// 通道关闭后停止监听并保留最后一次文本。messages[0] 作为初始文本，不能为空。
func WithUpdates(ch <-chan string) Option {
	return func(s *Spinner) {
		s.updates = ch
	}
}

type Spinner struct {
	writer io.Writer

	messages []string
	updates  <-chan string
	style    Style

	charInterval time.Duration
	holdDuration time.Duration
	frameRate    time.Duration

	stop chan struct{}
	wg   sync.WaitGroup
	once sync.Once
}

func New(messages []string, opts ...Option) *Spinner {
	// 两种模式都要求至少有一条初始文本，否则无内容可展示。
	if len(messages) == 0 {
		messages = []string{"初始化中..."}
	}

	s := &Spinner{
		writer:   os.Stdout,
		messages: messages,
		style:    StyleDots,

		charInterval: 50 * time.Millisecond,
		holdDuration: 500 * time.Millisecond,
		frameRate:    200 * time.Millisecond,

		stop: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Spinner) Start() {
	s.wg.Add(1)
	if s.updates != nil {
		go s.runDynamic()
	} else {
		go s.runStatic()
	}
}

// render 渲染一帧：spinner 字符 + 已展示的文本，并清除行尾残留。
func (s *Spinner) render(frame string, runes []rune, visible int) {
	// \033[K 清除光标到行尾的残留，避免上一帧较长文本留下尾巴，
	// 同时保证光标停在最后一个可见字符之后（打字机光标效果）。
	if visible >= len(runes) {
		fmt.Fprintf(s.writer, "\r%s %s...\033[K", frame, string(runes[:visible]))
	} else {
		fmt.Fprintf(s.writer, "\r%s %s\033[K", frame, string(runes[:visible]))
	}
}

// clear 回到行首并清除整行。
func (s *Spinner) clear() { fmt.Fprint(s.writer, "\r\033[K") }

// runStatic 按固定列表轮播：逐句打字、停留、切下一句、循环。
func (s *Spinner) runStatic() {
	defer s.wg.Done()

	frame, msg, visible := 0, 0, 0
	var holdStart time.Time
	holding := false

	frames := s.style.Frames
	frameTicker := time.NewTicker(s.frameRate)
	defer frameTicker.Stop()

	charTicker := time.NewTicker(s.charInterval)
	defer charTicker.Stop()

	runes := []rune(s.messages[msg])

	for {
		select {
		case <-s.stop:
			s.clear()
			return

		case <-frameTicker.C:
			// 仅更新 frame 索引并重新渲染
			frame = (frame + 1) % len(frames)
			s.render(frames[frame], runes, visible)

		case <-charTicker.C:
			// 处理字符逐字显示逻辑
			if !holding {
				if visible < len(runes) {
					visible++
				}
				if visible >= len(runes) {
					holding = true
					holdStart = time.Now()
				}
			} else if time.Since(holdStart) >= s.holdDuration {
				// 切换到下一条消息
				msg = (msg + 1) % len(s.messages)
				runes = []rune(s.messages[msg])
				visible = 0
				holding = false
			}
			s.render(frames[frame], runes, visible)
		}
	}
}

func commonPrefixLen(a, b string) int {
	ar := []rune(a)
	br := []rune(b)

	n := min(len(ar), len(br))

	i := 0
	for i < n && ar[i] == br[i] {
		i++
	}

	return i
}

// runDynamic 由外部通道驱动：收到新文本则重新逐字展示；无新文本时保持当前文本，
// spinner 动画持续。通道关闭后停止监听并保留最后一次文本。
func (s *Spinner) runDynamic() {
	defer s.wg.Done()

	frame, visible := 0, 0

	// messages[0] 作为初始文本（Start 已保证非空）。
	runes := []rune(s.messages[0])

	updates := s.updates
	frames := s.style.Frames
	frameTicker := time.NewTicker(s.frameRate)
	defer frameTicker.Stop()

	charTicker := time.NewTicker(s.charInterval)
	defer charTicker.Stop()

	for {
		select {
		case <-s.stop:
			s.clear()
			return

		case text, ok := <-updates:
			if !ok {
				// 通道关闭：停止监听，保留当前文本继续动画。
				updates = nil
				continue
			}

			// 取最新文本，重置为逐字展示。
			cpl := commonPrefixLen(string(runes), text)
			visible = min(visible, cpl)
			runes = []rune(text)

		case <-frameTicker.C:
			// 仅更新 frame 索引并重新渲染
			frame = (frame + 1) % len(frames)
			s.render(frames[frame], runes, visible)

		case <-charTicker.C:
			// 处理字符逐字显示逻辑
			if visible < len(runes) {
				visible++
				s.render(frames[frame], runes, visible)
			}
		}
	}
}

func (s *Spinner) Stop() {
	s.once.Do(func() {
		close(s.stop)
	})

	s.wg.Wait()
}
