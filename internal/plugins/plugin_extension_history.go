package plugins

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hpcloud/tail"
	"github.com/ruandada/aish/internal/base"
)

func (p *ExtensionPlugin) handleHistoryCommand(sce *base.SubCommandExecution, cmd string, args []string) error {
	commandLine := flag.NewFlagSet(cmd, flag.ContinueOnError)
	commandLine.SetOutput(sce.Stderr())
	commandLine.Usage = func() {
		fmt.Fprint(commandLine.Output(), "Usage:\n  history [n=100]\n\n")
		commandLine.PrintDefaults()
	}

	commandLine.Parse(args)
	args = commandLine.Args()

	n := 100
	if len(args) > 0 && args[0] != "" {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			commandLine.Usage()
			return nil
		}
		n = num
	}

	historyFile := filepath.Join(p.shell.State().User().HomeDir, base.HistoryFileName)
	if _, err := os.Stat(historyFile); err != nil {
		return nil
	}

	t, err := tail.TailFile(historyFile, tail.Config{
		Follow: false,
		Logger: tail.DiscardingLogger,
	})
	if err != nil {
		return err
	}
	defer t.Cleanup()
	defer t.Stop()

	stdout := sce.Stdout()
	i := 0
	for line := range t.Lines {
		stdout.Write([]byte(line.Text))
		stdout.Write([]byte("\n"))
		if i >= n {
			break
		}
		i++
	}

	return nil
}
