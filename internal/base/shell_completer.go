package base

import (
	"github.com/chzyer/readline"
)

type ShellCompleter struct {
	shell *Shell
}

func NewShellCompleter(shell *Shell) *ShellCompleter {
	return &ShellCompleter{
		shell: shell,
	}
}

var _ readline.AutoCompleter = &ShellCompleter{}

// Do implements readline.AutoCompleter.
func (s *ShellCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	if len(line) == 0 {
		return [][]rune{}, 0
	}

	for _, plugin := range s.shell.plugins {
		if c, l := plugin.AutoComplete(line, pos, s.shell); len(c) > 0 {
			return c, l
		}
	}

	return [][]rune{}, 0
}
