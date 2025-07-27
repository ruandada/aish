package plugins

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ruandada/aish/internal/base"
)

func (p *ExtensionPlugin) handleAIPromptCommand(sce *base.SubCommandExecution, cmd string, args []string) error {
	commandLine := flag.NewFlagSet(cmd, flag.ContinueOnError)
	commandLine.SetOutput(sce.Stderr())
	commandLine.Usage = func() {
		fmt.Fprint(commandLine.Output(), "Usage:\n  aiprompt\n  aiprompt \"<prompt1> <prompt2> ...\"\n  aiprompt reset\n\n")
		commandLine.PrintDefaults()
	}

	err := commandLine.Parse(args)
	if err != nil {
		return err
	}

	args = commandLine.Args()
	switch len(args) {
	case 0:
		sce.Stdout().Write([]byte(base.GetDefinedSystemPrompts()))
		sce.Stdout().Write([]byte("\n"))
	case 1:
		if prompt := strings.TrimSpace(args[0]); prompt == "reset" {
			base.ClearDefinedSystemPrompts()
		} else if prompt != "" {
			base.AddDefinedSystemPrompt(prompt)
		}

	default:
		for _, prompt := range args {
			prompt = strings.TrimSpace(prompt)
			if prompt == "" {
				continue
			}
			base.AddDefinedSystemPrompt(prompt)
		}
	}
	return nil
}
