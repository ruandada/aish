package plugins

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ruandada/aish/internal/base"
)

// Unicode Icons
const (
	IconUser  = "🚀"
	IconAuto  = "🪄"
	IconAI    = "💬"
	IconArrow = "➜"
)

type PromptPlugin struct {
}

func (p *PromptPlugin) Install(shell *base.Shell) error {
	return nil
}

var _ base.ShellPlugin = (*PromptPlugin)(nil)

func NewPromptPlugin() *PromptPlugin {
	return &PromptPlugin{}
}

func (p *PromptPlugin) ID() string {
	return "context"
}

// GeneratePrompt implements base.ShellPlugin.
func (p *PromptPlugin) GeneratePrompt(ce *base.CommandExecution, shell *base.Shell) (ok bool, prompt string, err error) {
	if !ce.Interactive() {
		return true, "", nil
	}

	if ce.Incomplete() {
		return true, "> ", nil
	}

	state := shell.State()
	mode := state.Mode()
	wd := shell.Dir()

	var modeIcon, modeText, modeColor string
	switch mode {
	case base.ShellModeAuto:
		modeColor = base.ColorPurple
		modeIcon = IconAuto
		modeText = "AUTO"
	case base.ShellModeUser:
		modeColor = base.ColorGreen
		modeIcon = IconUser
		modeText = "USER"
	case base.ShellModeAI:
		modeColor = base.ColorBlue
		modeIcon = IconAI
		modeText = "AI"
	}

	timeString := time.Now().Format("15:04")

	dirName := filepath.Base(wd)
	if wd == state.User().HomeDir {
		dirName = "~"
	}

	if !ce.ColorSupported() {
		return true, fmt.Sprintf(
			"[%s] %s %s %s ",
			modeText,
			dirName,
			timeString,
			IconArrow,
		), nil
	}

	// build colorful prompt
	parts := []string{
		fmt.Sprintf("%s%s %s%s", modeColor, modeIcon, modeText, base.ColorReset),
	}

	// 用户和目录
	parts = append(parts, fmt.Sprintf("%s%s%s%s", base.ColorCyan, base.Bold, dirName, base.ColorReset))

	parts = append(parts, fmt.Sprintf("%s%s%s", base.ColorGray, timeString, base.ColorReset))
	parts = append(parts, fmt.Sprintf("%s%s%s ", modeColor, IconArrow, base.ColorReset))

	return true, strings.Join(parts, " "), nil
}

func (a *PromptPlugin) PrepareContext(ce *base.CommandExecution, shell *base.Shell) (context.Context, error) {
	return nil, nil
}

func (p *PromptPlugin) BeforeExecute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) error {
	return nil
}

func (p *PromptPlugin) Execute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) (ok bool, err error) {
	return false, nil
}

func (p *PromptPlugin) AfterExecute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) error {
	return nil
}

// AutoComplete implements base.ShellPlugin.
func (a *PromptPlugin) AutoComplete(line []rune, pos int, shell *base.Shell) (newLine [][]rune, length int) {
	return nil, 0
}

func (p *PromptPlugin) End(ce *base.CommandExecution, shell *base.Shell) error {
	return nil
}
