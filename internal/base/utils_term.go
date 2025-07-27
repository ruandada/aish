package base

import (
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

const DefaultFileName = "aish"

// IsColorSupported checks if the terminal supports color.
func isColorSupported() bool {
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	// 常见的支持颜色的终端类型
	colorTerms := []string{"xterm", "xterm-256color", "screen", "tmux", "rxvt"}
	for _, ct := range colorTerms {
		if strings.Contains(term, ct) {
			return true
		}
	}

	return os.Getenv("COLORTERM") != ""
}

var colorSupported = isColorSupported()

// ANSI color code
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"

	// BG colors
	BgBlue   = "\033[44m"
	BgGreen  = "\033[42m"
	BgRed    = "\033[41m"
	BgYellow = "\033[43m"

	// 样式
	Bold      = "\033[1m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
)

func IsInteractive(writer io.Writer) bool {
	if writer == nil {
		return false
	}

	if f, ok := writer.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
