package base

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

var ErrPanic = errors.New("panic")

type modifierFuncKey struct{}

type Shell struct {
	stdin            *os.File
	stdout           *os.File
	stderr           *os.File
	plugins          []ShellPlugin
	state            *ShellState
	killTimeout      time.Duration
	fileName         string
	absoluteFileName string

	capturedStdout *os.File
	capturedStderr *os.File

	runner  *interp.Runner
	environ []string
	params  []string
	exit    bool
}

type ShellOption func(*Shell)

func WithStdIO(stdin *os.File, stdout *os.File, stderr *os.File) ShellOption {
	return func(s *Shell) {
		s.stdin = stdin
		s.stdout = stdout
		s.stderr = stderr
	}
}

func WithEnviron(environ []string) ShellOption {
	return func(s *Shell) {
		s.environ = environ
	}
}

func WithParams(params []string) ShellOption {
	return func(s *Shell) {
		s.params = params
	}
}

func WithFileName(fileName string, absoluteFileName string) ShellOption {
	return func(s *Shell) {
		s.fileName = fileName
		s.absoluteFileName = absoluteFileName
	}
}

func NewShell(opts ...ShellOption) (*Shell, error) {
	state, err := NewDefaultShellState()
	if err != nil {
		return nil, err
	}

	s := &Shell{
		stdin:       os.Stdin,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		plugins:     make([]ShellPlugin, 0, 8),
		state:       state,
		killTimeout: 2 * time.Second,
	}

	for _, opt := range opts {
		opt(s)
	}

	if capturedStdout, err := NewCapturedStdIO(s, s.stdout); err == nil {
		s.capturedStdout = capturedStdout
	} else {
		s.capturedStderr = s.stdout
	}
	if capturedStderr, err := NewCapturedStdIO(s, s.stderr); err == nil {
		s.capturedStderr = capturedStderr
	} else {
		s.capturedStderr = s.stderr
	}

	if s.fileName == "" {
		s.fileName = DefaultFileName
	}

	runnerOpts := []interp.RunnerOption{
		interp.Interactive(true),
		interp.StdIO(s.stdin, s.stdout, s.stderr),
		interp.ExecHandlers(func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
			return func(ctx context.Context, args []string) error {
				return s.execHandler(ctx, args)
			}
		}),
		interp.Params(s.params...),
	}

	environ := os.Environ()
	if len(s.environ) > 0 {
		environ = append(environ, s.environ...)
	}
	runnerOpts = append(runnerOpts, interp.Env(expand.ListEnviron(environ...)))

	r, err := interp.New(runnerOpts...)
	if err != nil {
		return nil, err
	}
	s.runner = r
	return s, nil
}

func (s *Shell) Use(plugin ...ShellPlugin) error {
	for _, p := range plugin {
		if err := p.Install(s); err != nil {
			return err
		}
		s.plugins = append(s.plugins, p)
	}
	return nil
}

func (s *Shell) State() *ShellState {
	return s.state
}

func (s *Shell) readlines(
	ctx context.Context,
	stdin *os.File,
) error {
	isTerminal := IsInteractive(stdin)

	plugins := s.plugins
	eof := make(chan struct{})
	readlineStart := make(chan *CommandExecution)
	readlineEnd := make(chan []byte)
	defer close(readlineStart)

	var interactiveReader *readline.Instance
	var nonInteractiveReader *bufio.Reader

	if isTerminal {
		if rl, err := readline.NewEx(&readline.Config{
			Prompt:       "",
			HistoryFile:  filepath.Join(s.state.User().HomeDir, HistoryFileName),
			Stdin:        stdin,
			Stdout:       s.capturedStdout,
			Stderr:       s.capturedStderr,
			AutoComplete: NewShellCompleter(s),
		}); err == nil {
			interactiveReader = rl
			defer rl.Close()
		} else {
			return err
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan)
		defer func() {
			signal.Stop(sigChan)
			close(sigChan)
		}()

		go func() {
			for sig := range sigChan {
				s.processSignal(sig)
			}
		}()
	} else {
		nonInteractiveReader = bufio.NewReader(stdin)
	}

	go func() {
		readNextLine := func(ce *CommandExecution) (string, error) {
			var line string
			var err error

			for {
				if interactiveReader != nil {
					for _, plugin := range plugins {
						if ok, prompt, err := plugin.GeneratePrompt(ce, s); ok {
							if err != nil {
								s.PrintError(s.stderr, err)
							} else {
								interactiveReader.SetPrompt(prompt)
							}
							break
						}
					}

					line, err = interactiveReader.Readline()
				} else if nonInteractiveReader != nil {
					line, err = nonInteractiveReader.ReadString('\n')
				} else {
					return "", errors.New("no reader")
				}
				if err != nil {
					if err == readline.ErrInterrupt {
						return line, context.Canceled
					}
					return line, err
				}

				if strings.TrimSpace(line) == "" {
					continue
				}

				return line, err
			}
		}

		for {
			ce, ok := <-readlineStart
			if !ok {
				break
			}

			line, err := readNextLine(ce)
			if err != nil {
				if err == context.Canceled {
					ce.Cancel()
					continue
				} else if err == io.EOF {
					if line != "" {
						readlineEnd <- []byte(line)
					} else {
						readlineEnd <- nil
					}
					break
				} else {
					readlineEnd <- nil
					continue
				}
			}

			readlineEnd <- []byte(line)
		}
		close(readlineEnd)
		close(eof)
	}()

	for {
		ce := s.newCommandExecution(ctx, isTerminal)
		parser := syntax.NewParser()
		ast := &syntax.File{
			Name: s.fileName,
		}
		cio := NewChanIO(1)

		waitNextLine := func() error {
			ctx := ce.Context()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-eof:
				return io.EOF
			case readlineStart <- ce:
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-eof:
					return io.EOF
				case code, ok := <-readlineEnd:
					if !ok {
						return io.EOF
					}
					cio.Write(append(code, '\n'))
				}
			}
			return nil
		}

		interrupt := false
		if err := waitNextLine(); err != nil {
			if err == context.Canceled {
				interrupt = true
			} else {
				if err != io.EOF {
					s.PrintError(s.stderr, err)
				}
				break
			}
		}

		if interrupt {
			ce.Cancel()
			continue
		}

		err := parser.Interactive(cio, func(stmts []*syntax.Stmt) bool {
			if parser.Incomplete() {
				ce.incomplete = true
				if err := waitNextLine(); err != nil {
					if err == context.Canceled {
						interrupt = true
					} else if err != io.EOF {
						s.PrintError(s.stderr, err)
					}
					return false
				}
				return true
			}
			ce.incomplete = false
			ast.Stmts = stmts
			return false
		})
		cio.Close()

		if err != nil {
			if !interrupt {
				s.PrintError(s.stderr, err)
			}
			continue
		}

		s.state.SetCurrentExecution(ce)
		for _, plugin := range s.plugins {
			if c, err := plugin.PrepareContext(ce, s); err != nil {
				return err
			} else if c != nil {
				ce.ctx = c
			}
		}

		if err := s.evalAST(ce, ast, nil); err != nil {
			if err == ErrPanic {
				err = s.handlePanic(ce, err, cio.Bytes())
			}

			if err != nil {
				s.PrintError(s.stderr, err)
			}
		}

		for _, plugin := range s.plugins {
			if err := plugin.End(ce, s); err != nil {
				s.PrintError(s.stderr, err)
			}
		}
		ce.terminated = true
		s.state.SetCurrentExecution(nil)

		if s.exit {
			break
		}
	}
	return nil
}

func (s *Shell) Start(ctx context.Context) error {
	if home, err := os.UserHomeDir(); err == nil {
		s.readWorkspaceConfig(ctx, home)
	}

	if wd, err := os.Getwd(); err == nil {
		s.readWorkspaceConfig(ctx, wd)
	}

	return s.readlines(ctx, s.stdin)
}

func (s *Shell) evalAST(ce *CommandExecution, ast *syntax.File, modifierFunc func(sce *SubCommandExecution)) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = ErrPanic
		}
	}()

	ast.Name = s.fileName
	runnerCtx := context.WithValue(ce.parentCtx, commandExecutionKey{}, ce)
	if modifierFunc != nil {
		runnerCtx = context.WithValue(runnerCtx, modifierFuncKey{}, modifierFunc)
	}
	err = s.runner.Run(runnerCtx, ast)
	if s.runner.Exited() {
		s.exit = true
		return err
	}

	if err != nil {
		return err
	}
	return nil
}

func (s *Shell) Eval(ce *CommandExecution, code []byte, modifierFunc func(sce *SubCommandExecution)) error {
	if len(code) == 0 {
		return nil
	}

	ast, err := syntax.NewParser().Parse(bytes.NewReader(code), "")
	if err != nil {
		return err
	}

	return s.evalAST(ce, ast, modifierFunc)
}

func (s *Shell) processSignal(sig os.Signal) {
	switch sig {
	case syscall.SIGINT:
		if ce := s.state.CurrentExecution(); ce != nil {
			ce.Cancel()
		}
	}
}

func (s *Shell) Dir() string {
	return s.runner.Dir
}

func (s *Shell) LookPath(file string) (string, error) {
	return LookPath(file, s.Dir(), s.runner.Env.Get("PATH").String())
}

func (s *Shell) FindExecutableNames() ([]string, error) {
	return FindExecutableNames(s.runner.Env.Get("PATH").String(), s.Dir())
}

func (s *Shell) AbsoluteFileName() string {
	return s.absoluteFileName
}

func (s *Shell) FileName() string {
	return s.fileName
}
