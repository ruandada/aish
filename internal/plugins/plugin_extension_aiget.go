package plugins

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/ruandada/aish/internal/base"
)

func (p *ExtensionPlugin) handleAIGetCommand(sce *base.SubCommandExecution, cmd string, args []string) error {
	commandLine := flag.NewFlagSet(cmd, flag.ContinueOnError)
	commandLine.SetOutput(sce.Stderr())
	commandLine.Usage = func() {
		fmt.Fprint(commandLine.Output(), "Usage:\n  aiget <key>\n\n")
		commandLine.PrintDefaults()
	}

	err := commandLine.Parse(args)
	if err != nil {
		return err
	}

	args = commandLine.Args()
	switch len(args) {
	case 0:
		cfg := base.GetAllConfig()
		for k, v := range cfg {
			fmt.Fprintf(sce.Stdout(), "%s=%s\n", k, v)
		}
		return nil
	case 1:
		value := base.GetConfig(base.ConfigName(args[0]))
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}
		sce.Stdout().Write(b)
		sce.Stdout().Write([]byte("\n"))
		return nil
	default:
		commandLine.Usage()
		return nil
	}
}
