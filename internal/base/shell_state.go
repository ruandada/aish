package base

import (
	"os/user"
	"runtime"
	"sync"
)

type ShellMode int

const (
	ShellModeAuto ShellMode = iota
	ShellModeUser ShellMode = 1
	ShellModeAI   ShellMode = 2
)

type ShellState struct {
	os               string
	arch             string
	user             *user.User
	mode             ShellMode
	currentExecution *CommandExecution
	mu               sync.RWMutex
}

func NewDefaultShellState() (*ShellState, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	return &ShellState{
		os:   runtime.GOOS,
		arch: runtime.GOARCH,
		user: user,
		mode: ShellModeAuto,
	}, nil
}

func (s *ShellState) User() *user.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.user
}

func (s *ShellState) Mode() ShellMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

func (s *ShellState) SetMode(mode ShellMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
}

func (s *ShellState) CurrentExecution() *CommandExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentExecution
}

func (s *ShellState) SetCurrentExecution(ce *CommandExecution) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentExecution = ce
}

func (s *ShellState) OS() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.os
}

func (s *ShellState) Arch() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.arch
}
