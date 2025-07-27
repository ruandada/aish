package base

import (
	"context"
	"os"
	"path/filepath"
)

const (
	RCFileName      = ".aishrc"
	HistoryFileName = ".aish_history"
)

func (s *Shell) readWorkspaceConfig(ctx context.Context, workspace string) {
	rcFile := filepath.Join(workspace, RCFileName)

	file, err := os.Open(rcFile)
	if err != nil {
		if !os.IsNotExist(err) {
			s.PrintError(s.stderr, err)
		}
		return
	}
	defer file.Close()

	if err := s.readlines(ctx, file); err != nil {
		s.PrintError(s.stderr, err)
	}
}
