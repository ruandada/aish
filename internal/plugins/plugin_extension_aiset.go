package plugins

import (
	"flag"
	"fmt"

	"github.com/ruandada/aish/internal/base"
)

func (p *ExtensionPlugin) handleAISetCommand(sce *base.SubCommandExecution, cmd string, args []string) error {
	commandLine := flag.NewFlagSet(cmd, flag.ContinueOnError)
	commandLine.SetOutput(sce.Stderr())
	commandLine.Usage = func() {
		fmt.Fprint(commandLine.Output(), "Usage:\n  aiset <key> <value>\n\n")
		commandLine.PrintDefaults()
	}

	err := commandLine.Parse(args)
	if err != nil {
		return err
	}

	args = commandLine.Args()
	if len(args) != 2 {
		commandLine.Usage()
		return nil
	}

	key := args[0]
	value := args[1]
	base.SetConfig(base.ConfigName(key), value)
	return nil
}
