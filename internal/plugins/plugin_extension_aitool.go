package plugins

import (
	"flag"
	"fmt"

	"github.com/ruandada/aish/internal/base"
)

func (p *ExtensionPlugin) handleAIToolCommand(sce *base.SubCommandExecution, cmd string, args []string) error {
	commandLine := flag.NewFlagSet(cmd, flag.ContinueOnError)
	commandLine.SetOutput(sce.Stderr())
	commandLine.Usage = func() {
		fmt.Fprint(commandLine.Output(), "Usage:\n  aitool -u \"<usage>\" <entrypoint>\n  aitool clear\n\n")
		commandLine.PrintDefaults()
	}

	usage := ""
	commandLine.StringVar(&usage, "u", "", "short text describing usage of this tool")

	err := commandLine.Parse(args)
	if err != nil {
		return err
	}

	args = commandLine.Args()
	switch len(args) {
	case 0:
		sce.Stdout().Write([]byte("AI tools:\n\n"))
		tools := base.GetDefinedTools()
		for _, tool := range tools {
			fmt.Fprintf(sce.Stdout(), "%s\n[usage=%s]\n\n", tool.Entrypoint, tool.Usage)
		}
		return nil
	case 1:
		entrypoint := args[0]
		if entrypoint == "clear" {
			base.ClearDefinedTools()
			return nil
		}

		filepath, err := p.shell.LookPath(entrypoint)
		if err != nil {
			return err
		}

		if filepath == p.shell.AbsoluteFileName() {
			// currently is running tool file, skip
			return nil
		}

		if err := base.AddDefinedTool(usage, filepath); err != nil {
			return err
		}
	default:
		commandLine.Usage()
		return nil
	}

	return nil
}
