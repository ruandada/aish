package base

import (
	"context"
)

type ShellPlugin interface {
	ID() string

	// A Shell may execute multiple CommandExecution, and each CommandExecution may execute multiple SubCommandExecution.

	// Shell related
	Install(shell *Shell) error
	AutoComplete(line []rune, pos int, shell *Shell) (newLine [][]rune, length int)

	// CommandExecution related
	PrepareContext(ce *CommandExecution, shell *Shell) (context.Context, error)
	GeneratePrompt(ce *CommandExecution, shell *Shell) (ok bool, prompt string, err error)
	End(ce *CommandExecution, shell *Shell) error

	// SubCommandExecution related
	BeforeExecute(ce *CommandExecution, sce *SubCommandExecution, shell *Shell) error
	Execute(ce *CommandExecution, sce *SubCommandExecution, shell *Shell) (ok bool, err error)
	AfterExecute(ce *CommandExecution, sce *SubCommandExecution, shell *Shell) error
}
