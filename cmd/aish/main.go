package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ruandada/aish/internal/base"
	"github.com/ruandada/aish/internal/plugins"
)

var (
	command = flag.String("c", "", "command to execute")
)

func handleError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
	os.Exit(1)
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		handleError(err)
		return
	}

	executable, err := os.Executable()
	if err != nil {
		handleError(err)
		return
	}
	environ := []string{
		fmt.Sprintf("SHELL=%s", executable),
		fmt.Sprintf("AISH=%s", executable),
	}

	flag.Parse()
	args := flag.Args()
	var params []string

	filename, absoluteFileName := "", ""
	if n := len(args); n > 0 {
		filename = args[0]
		if n > 1 {
			params = args[1:]
		}
	}

	var stdin *os.File

	switch {
	// if user use "-c" to execute an inline command, use it as stdin
	case command != nil && *command != "":
		file, err := base.ReaderDescriptor(strings.NewReader(*command))
		if err != nil {
			handleError(err)
			return
		}
		defer file.Close()
		stdin = file

	case filename != "":
		abs, err := base.LookPath(filename, wd, os.Getenv("PATH"))
		if err != nil {
			handleError(err)
			return
		}
		absoluteFileName = abs
		file, err := os.Open(abs)
		if err != nil {
			handleError(err)
			return
		}
		defer file.Close()
		stdin = file

	default:
		stdin = os.Stdin
	}

	shell, err := base.NewShell(
		base.WithStdIO(stdin, os.Stdout, os.Stderr),
		base.WithParams(params),
		base.WithFileName(filename, absoluteFileName),
		base.WithEnviron(environ),
	)
	if err != nil {
		handleError(err)
		return
	}

	if err := shell.Use(
		plugins.NewPromptPlugin(),
		plugins.NewPathAutocompletePlugin(),
		plugins.NewExtensionPlugin(),
		plugins.NewAIPlugin(),
	); err != nil {
		shell.PrintError(os.Stderr, err)
		os.Exit(1)
	}

	if err := shell.Start(context.Background()); err != nil {
		shell.PrintError(os.Stderr, err)
		os.Exit(1)
	}
}
