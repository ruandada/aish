package base

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/iancoleman/strcase"
)

type DefinedTool struct {
	Name       string
	Usage      string
	Entrypoint string
}

var definedTools = make(map[string]*DefinedTool)

func AddDefinedTool(usage string, entrypoint string) error {
	if _, err := os.Stat(entrypoint); err != nil {
		return err
	}

	toolName := strcase.ToSnake(filepath.Base(entrypoint))
	if _, ok := definedTools[toolName]; ok {
		return fmt.Errorf("%s: tool %s already defined", entrypoint, toolName)
	}

	definedTools[toolName] = &DefinedTool{
		Usage:      usage,
		Name:       toolName,
		Entrypoint: entrypoint,
	}
	return nil
}

func ClearDefinedTools() {
	definedTools = make(map[string]*DefinedTool)
}

func GetDefinedTools() map[string]*DefinedTool {
	return definedTools
}

func GetDefinedTool(name string) (tool *DefinedTool, ok bool) {
	t, ok := definedTools[name]
	return t, ok
}
