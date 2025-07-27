package plugins

import (
	"context"
	"strings"

	"github.com/chzyer/readline"
	"github.com/ruandada/aish/internal/base"
)

type ExtensionCommandName string

const (
	ExtensionCommandAutoMode      ExtensionCommandName = "auto:"
	ExtensionCommandAIMode        ExtensionCommandName = "ai:"
	ExtensionCommandUserMode      ExtensionCommandName = "user:"
	ExtensionCommandUserModeShort ExtensionCommandName = "::"
	ExtensionCommandAISet         ExtensionCommandName = "aiset"
	ExtensionCommandAIGet         ExtensionCommandName = "aiget"
	ExtensionCommandAIPrompt      ExtensionCommandName = "aiprompt"
	ExtensionCommandAITool        ExtensionCommandName = "aitool"
	ExtensionCommandHistory       ExtensionCommandName = "history"
)

var builtinCommands = []string{
	":",
	"true",
	"false",
	"exit",
	"set",
	"shift",
	"unset",
	"echo",
	"printf",
	"break",
	"continue",
	"pwd",
	"cd",
	"wait",
	"builtin",
	"trap",
	"type",
	"source",
	".",
	"command",
	"return",
	"read",
	"mapfile",
	"readarray",
	"shopt",
}

type ExtensionPlugin struct {
	shell          *base.Shell
	pathCommands   []string
	prefxCompleter readline.PrefixCompleterInterface
}

var _ base.ShellPlugin = (*ExtensionPlugin)(nil)

func NewExtensionPlugin() *ExtensionPlugin {
	return &ExtensionPlugin{}
}

// ID implements base.ShellPlugin.
func (p *ExtensionPlugin) ID() string {
	return "extension"
}

func (p *ExtensionPlugin) BeforeExecute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) error {
	return nil
}

// Execute implements base.ShellPlugin.
func (p *ExtensionPlugin) Execute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) (ok bool, err error) {
	fields := sce.Fields()
	if len(fields) == 0 {
		return false, nil
	}
	cmd, args := strings.ToLower(fields[0]), fields[1:]

	switch cmd {
	case string(ExtensionCommandAutoMode):
		if done := p.handleModeSwitch(sce, base.ShellModeAuto, args); done {
			return true, nil
		}
	case string(ExtensionCommandUserModeShort):
		fallthrough
	case string(ExtensionCommandUserMode):
		if done := p.handleModeSwitch(sce, base.ShellModeUser, args); done {
			return true, nil
		}
	case string(ExtensionCommandAIMode):
		if done := p.handleModeSwitch(sce, base.ShellModeAI, args); done {
			return true, nil
		}
	}

	fields = sce.Fields()
	cmd, args = strings.ToLower(fields[0]), fields[1:]

	switch cmd {
	case string(ExtensionCommandAISet):
		if err := p.handleAISetCommand(sce, cmd, args); err != nil {
			shell.PrintError(sce.Stderr(), err)
		}
		return true, nil
	case string(ExtensionCommandAIGet):
		if err := p.handleAIGetCommand(sce, cmd, args); err != nil {
			shell.PrintError(sce.Stderr(), err)
		}
		return true, nil
	case string(ExtensionCommandAIPrompt):
		if err := p.handleAIPromptCommand(sce, cmd, args); err != nil {
			shell.PrintError(sce.Stderr(), err)
		}
		return true, nil
	case string(ExtensionCommandAITool):
		if err := p.handleAIToolCommand(sce, cmd, args); err != nil {
			shell.PrintError(sce.Stderr(), err)
		}
		return true, nil
	case string(ExtensionCommandHistory):
		if err := p.handleHistoryCommand(sce, cmd, args); err != nil {
			shell.PrintError(sce.Stderr(), err)
		}
		return true, nil
	default:
		return false, nil
	}
}

func (p *ExtensionPlugin) handleModeSwitch(sce *base.SubCommandExecution, mode base.ShellMode, args []string) (done bool) {
	if len(args) == 0 {
		p.shell.State().SetMode(mode)
		return true
	} else {
		sce.SetMode(mode)
		sce.SetFields(args)
	}
	return false
}

// PrepareContext implements base.ShellPlugin.
func (p *ExtensionPlugin) PrepareContext(ce *base.CommandExecution, shell *base.Shell) (context.Context, error) {
	return nil, nil
}

// AutoComplete implements base.ShellPlugin.
func (p *ExtensionPlugin) AutoComplete(line []rune, pos int, shell *base.Shell) (newLine [][]rune, length int) {
	return p.prefxCompleter.Do(line, pos)
}

func (p *ExtensionPlugin) rebuildAutoCompleter() {
	configItems := make([]readline.PrefixCompleterInterface, 0, len(base.ConfigKeys))
	for _, key := range base.ConfigKeys {
		configItems = append(configItems, readline.PcItem(string(key)))
	}

	commandCompleters := []readline.PrefixCompleterInterface{
		readline.PcItem(string(ExtensionCommandAISet), configItems...),
		readline.PcItem(string(ExtensionCommandAIGet), configItems...),
		readline.PcItem(string(ExtensionCommandAIPrompt), readline.PcItem("clear")),
		readline.PcItem(string(ExtensionCommandAITool), readline.PcItem("clear")),
	}

	for _, cmd := range builtinCommands {
		commandCompleters = append(commandCompleters, readline.PcItem(cmd))
	}

	if p.pathCommands != nil {
		for _, executable := range p.pathCommands {
			commandCompleters = append(commandCompleters, readline.PcItem(executable))
		}
	}

	completers := append([]readline.PrefixCompleterInterface{
		readline.PcItem("ai:"),
		readline.PcItem("user:", commandCompleters...),
		readline.PcItem("::", commandCompleters...),
	}, commandCompleters...)

	p.prefxCompleter = readline.NewPrefixCompleter(completers...)
}

func (p *ExtensionPlugin) Install(shell *base.Shell) error {
	p.shell = shell

	p.rebuildAutoCompleter()

	go func() {
		pathCommands, err := shell.FindExecutableNames()
		if err != nil {
			return
		}
		p.pathCommands = pathCommands
		p.rebuildAutoCompleter()
	}()
	return nil
}

// GeneratePrompt implements base.ShellPlugin.
func (p *ExtensionPlugin) GeneratePrompt(ce *base.CommandExecution, shell *base.Shell) (ok bool, prompt string, err error) {
	return false, "", nil
}

// AfterExecute implements base.ShellPlugin.
func (p *ExtensionPlugin) AfterExecute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) error {
	return nil
}

func (p *ExtensionPlugin) End(ce *base.CommandExecution, shell *base.Shell) error {
	return nil
}
