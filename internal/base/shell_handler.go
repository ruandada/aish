package base

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func (s *Shell) execHandler(ctx context.Context, args []string) error {
	ce, ok := GetCommandExecution(ctx)
	if !ok {
		return errors.New("command execution not detected")
	}

	hc := interp.HandlerCtx(ctx)
	sce := ce.NewSubCommandExecution(args, &hc)

	if modifierFunc, ok := ctx.Value(modifierFuncKey{}).(func(sce *SubCommandExecution)); ok {
		modifierFunc(sce)
	}
	// capture the command output for AI context
	if hc.Stdout == s.stdout {
		hc.Stdout = s.capturedStdout
	}
	if hc.Stderr == s.stderr {
		hc.Stderr = s.capturedStderr
	}

	plugins := s.plugins

	for _, plugin := range plugins {
		if err := plugin.BeforeExecute(ce, sce, s); err != nil {
			sce.err = err
			return err
		}
	}

	done := false
	for _, plugin := range plugins {
		if ok, err := plugin.Execute(ce, sce, s); ok {
			if err != nil {
				sce.err = err
			}
			done = true
			break
		}
	}

	if !done {
		sce.err = errors.New("no executor")
		done = true
	}

	for _, plugin := range plugins {
		if err := plugin.AfterExecute(ce, sce, s); err != nil {
			sce.err = err
			return err
		}
	}

	return nil
}

func execEnv(env expand.Environ) []string {
	list := make([]string, 0, 64)
	for name, vr := range env.Each {
		if !vr.IsSet() {
			// If a variable is set globally but unset in the
			// runner, we need to ensure it's not part of the final
			// list. Seems like zeroing the element is enough.
			// This is a linear search, but this scenario should be
			// rare, and the number of variables shouldn't be large.
			for i, kv := range list {
				if strings.HasPrefix(kv, name+"=") {
					list[i] = ""
				}
			}
		}
		if vr.Exported && vr.Kind == expand.String {
			list = append(list, name+"="+vr.String())
		}
	}
	return list
}

func (s *Shell) handlePanic(ce *CommandExecution, err error, input []byte) error {
	if s.state.Mode() == ShellModeUser {
		return err
	}

	// If panic occurs, as fallback, we use the whole user input as a command to interact with AI
	ast := &syntax.File{
		Name: s.fileName,
		Stmts: []*syntax.Stmt{
			{
				Position: syntax.NewPos(0, 0, 0),
				Cmd: &syntax.CallExpr{
					Args: []*syntax.Word{
						{
							Parts: []syntax.WordPart{
								&syntax.DblQuoted{
									Parts: []syntax.WordPart{
										&syntax.Lit{
											Value: string(input),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return s.evalAST(ce, ast, nil)
}

func (sce *SubCommandExecution) DefaultExecHandler() error {
	ctx, shell, hc, args, mode := sce.ce.ctx, sce.ce.shell, sce.hc, sce.fields, sce.mode

	path, err := interp.LookPathDir(hc.Dir, hc.Env, args[0])
	if err != nil {
		if mode != ShellModeAuto {
			fmt.Fprintln(hc.Stderr, err)
		}
		return interp.ExitStatus(127)
	}
	cmd := exec.Cmd{
		Path:   path,
		Args:   args,
		Env:    execEnv(hc.Env),
		Dir:    hc.Dir,
		Stdin:  hc.Stdin,
		Stdout: hc.Stdout,
		Stderr: hc.Stderr,
	}

	err = cmd.Start()
	if err == nil {
		stopf := context.AfterFunc(ctx, func() {
			if shell.killTimeout <= 0 || runtime.GOOS == "windows" {
				_ = cmd.Process.Signal(os.Kill)
				return
			}
			_ = cmd.Process.Signal(os.Interrupt)
			// TODO: don't sleep in this goroutine if the program
			// stops itself with the interrupt above.
			time.Sleep(shell.killTimeout)
			_ = cmd.Process.Signal(os.Kill)
		})
		defer stopf()

		err = cmd.Wait()
	}

	switch err := err.(type) {
	case *exec.ExitError:
		// Windows and Plan9 do not have support for [syscall.WaitStatus]
		// with methods like Signaled and Signal, so for those, [waitStatus] is a no-op.
		// Note: [waitStatus] is an alias [syscall.WaitStatus]
		if status, ok := err.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return interp.ExitStatus(128 + status.Signal())
		}
		return interp.ExitStatus(err.ExitCode())
	case *exec.Error:
		// did not start
		if mode != ShellModeAuto {
			fmt.Fprintf(hc.Stderr, "%v\n", err)
		}
		return interp.ExitStatus(127)
	default:
		return err
	}
}
