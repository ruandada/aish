package base

import (
	"context"
	"io"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/openai/openai-go"
	"mvdan.cc/sh/v3/interp"
)

type CommandExecution struct {
	shell     *Shell
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc

	incomplete  bool
	interactive bool
	terminated  bool

	buf *strings.Builder

	qa []*AIExecution
}

type commandExecutionKey struct{}

func GetCommandExecution(ctx context.Context) (*CommandExecution, bool) {
	ce, ok := ctx.Value(commandExecutionKey{}).(*CommandExecution)
	if !ok || ce == nil {
		return nil, false
	}
	return ce, true
}

func (s *Shell) newCommandExecution(ctx context.Context, interactive bool) *CommandExecution {
	newCtx, cancel := context.WithCancel(ctx)

	ce := &CommandExecution{
		shell:       s,
		parentCtx:   ctx,
		ctx:         newCtx,
		cancel:      cancel,
		buf:         &strings.Builder{},
		interactive: interactive,
	}
	return ce
}

func (c *CommandExecution) AppendQA(qa *AIExecution) {
	c.qa = append(c.qa, qa)
}

func (c *CommandExecution) QA() []*AIExecution {
	return c.qa
}

func (c *CommandExecution) Context() context.Context {
	return c.ctx
}

func (c *CommandExecution) Cancel() {
	c.cancel()
}

func (c *CommandExecution) Buffer() *strings.Builder {
	return c.buf
}

func (c *CommandExecution) AnswerText() string {
	return strings.TrimSpace(stripansi.Strip(c.Buffer().String()))
}

func (c *CommandExecution) Interactive() bool {
	return c.interactive
}

func (c *CommandExecution) ColorSupported() bool {
	return colorSupported && c.interactive
}

func (c *CommandExecution) Terminated() bool {
	return c.terminated
}

func (c *CommandExecution) Incomplete() bool {
	return c.incomplete
}

type AIAssistantAnswer struct {
	Text     string                                `json:"text"`
	ToolCall *openai.ChatCompletionMessageToolCall `json:"tool_call"`
}

type AIExecution struct {
	parent        *AIExecution
	UnderToolCall *openai.ChatCompletionMessageToolCall
	Question      string
	Answers       []AIAssistantAnswer
}

func (e *AIExecution) IsRoot() bool {
	return e.parent == nil
}

func (e *AIExecution) Trace() []*AIExecution {
	cursor := e
	trace := []*AIExecution{}
	for cursor != nil {
		trace = append(trace, cursor)
		cursor = cursor.parent
	}
	return trace
}

type SubCommandExecution struct {
	mode        ShellMode
	ce          *CommandExecution
	fields      []string
	hc          *interp.HandlerContext
	err         error
	qa          *AIExecution
	interactive bool
	parent      *SubCommandExecution

	initialStdout io.Writer
	initialStderr io.Writer
}

func (ce *CommandExecution) NewSubCommandExecution(fields []string, hc *interp.HandlerContext) *SubCommandExecution {
	return &SubCommandExecution{
		ce:     ce,
		mode:   ce.shell.State().Mode(),
		fields: fields,
		hc:     hc,
		qa: &AIExecution{
			Question: "",
			Answers:  make([]AIAssistantAnswer, 0),
		},
		interactive: IsInteractive(hc.Stdout),

		initialStdout: hc.Stdout,
		initialStderr: hc.Stderr,
	}
}

func (c *SubCommandExecution) Fields() []string {
	return c.fields
}

func (c *SubCommandExecution) Cmd() string {
	if len(c.fields) == 0 {
		return ""
	}
	return c.fields[0]
}

func (c *SubCommandExecution) SetFields(fields []string) {
	c.fields = fields
}

func (c *SubCommandExecution) QA() *AIExecution {
	return c.qa
}

func (c *SubCommandExecution) Mode() ShellMode {
	return c.mode
}

func (c *SubCommandExecution) SetMode(mode ShellMode) {
	c.mode = mode
}

func (c *SubCommandExecution) Error() error {
	return c.err
}

// Used to write content that visible to both user and AI
func (c *SubCommandExecution) Stdout() io.Writer {
	if c.ce.terminated {
		return c.initialStdout
	}
	return c.hc.Stdout
}

// Used to write error content that visible to both user and AI
func (c *SubCommandExecution) Stderr() io.Writer {
	if c.ce.terminated {
		return c.initialStderr
	}
	return c.hc.Stderr
}

func (c *SubCommandExecution) Stdin() io.Reader {
	return c.hc.Stdin
}

// Used to write content that only visible to AI
func (c *SubCommandExecution) Stdai() io.Writer {
	return c.ce.Buffer()
}

func (c *SubCommandExecution) Interactive() bool {
	return c.interactive
}

func (c *SubCommandExecution) ColorSupported() bool {
	return colorSupported && c.interactive
}

func (c *SubCommandExecution) Inherit(parent *SubCommandExecution) {
	c.parent = parent
	c.qa.parent = parent.qa

	shell := c.ce.shell

	if c.hc.Stdin == shell.stdin {
		c.hc.Stdin = c.parent.hc.Stdin
	}
	if c.hc.Stdout == shell.stdout {
		c.hc.Stdout = c.parent.hc.Stdout
	}
	if c.hc.Stderr == shell.stderr {
		c.hc.Stderr = c.parent.hc.Stderr
	}
}
