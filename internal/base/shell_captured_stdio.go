package base

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
)

var sigwinch = make(chan os.Signal, 1)
var capturedIOs []*os.File = make([]*os.File, 0, 2)

func init() {
	if stdin := os.Stdin; IsInteractive(stdin) {
		signal.Notify(sigwinch, syscall.SIGWINCH)

		go func() {
			for range sigwinch {
				if len(capturedIOs) == 0 {
					continue
				}
				for _, f := range capturedIOs {
					_ = pty.InheritSize(stdin, f)
				}
			}
		}()
	}
}

type executionDualWriter struct {
	stdWriter io.Writer
	s         *Shell
}

var _ io.Writer = (*executionDualWriter)(nil)

func (w *executionDualWriter) Write(p []byte) (n int, err error) {
	n, err = w.stdWriter.Write(p)
	if err != nil {
		return n, err
	}

	// write stdout or stderr to the current execution buffer, which will be used to generate AI messages
	if ce := w.s.State().CurrentExecution(); ce != nil {
		return ce.Buffer().Write(p)
	}
	return len(p), nil
}

func NewCapturedStdIO(shell *Shell, writer io.Writer) (*os.File, error) {
	pr, pw, err := pty.Open()
	if err != nil {
		return nil, err
	}
	capturedIOs = append(capturedIOs, pw)
	sigwinch <- syscall.SIGWINCH

	go func() {
		io.Copy(
			&executionDualWriter{stdWriter: writer, s: shell},
			pr,
		)
		_ = pr.Close()
		_ = pw.Close()
	}()

	return pw, nil
}
