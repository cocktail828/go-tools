package spinner

// Style 定义 spinner 的动画展示风格：一组按帧率循环播放的字符。
type Style struct {
	Name   string
	Frames []string
}

// 内置展示风格，可通过 WithStyle 选择。
var (
	StyleDots    = Style{"dots", []string{"◐", "◓", "◑", "◒"}}
	StyleCircles = Style{"circles", []string{"●○○", "○●○", "○○●", "○●○"}}
	StyleStars   = Style{"stars", []string{"✳", "✴", "✷", "✸", "✷", "✴"}}
	StyleLine    = Style{"line", []string{"|", "/", "-", "\\"}}
	StyleBounce  = Style{"bounce", []string{"⠁", "⠂", "⠄", "⠂"}}
	StyleBraille = Style{"braille", []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}}
	StyleMoon    = Style{"moon", []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}}
	StyleClock   = Style{"clock", []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}}
	StyleGrow    = Style{"grow", []string{"▁", "▃", "▄", "▅", "▆", "▇", "▆", "▅", "▄", "▃"}}
	StylePulse   = Style{"pulse", []string{" ", "▌", "█", "▐", " ", "▐", "█", "▌"}}
)

// valid 判断风格是否可用（至少含一帧）。
func (s Style) valid() bool { return len(s.Frames) > 0 }
