package plugins

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"unicode"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/ruandada/aish/internal/base"
	"mvdan.cc/sh/v3/interp"
)

type ToolName string

const (
	ToolNameExecute           ToolName = "EXECUTE"
	ToolNameUserDefinedPrefix ToolName = "TOOL_"
)

//go:embed plugin_ai_system_prompt.tmpl
var systemPrompt []byte

var systemPromptTemplate *template.Template

func init() {
	t, err := template.New("system_prompt").Parse(string(systemPrompt))
	if err != nil {
		panic(err)
	}
	systemPromptTemplate = t
}

type AIPlugin struct {
	client            *openai.Client
	historyLimit      int
	historyExecutions []*base.AIExecution
	iterationLimit    int
}

var _ base.ShellPlugin = (*AIPlugin)(nil)

func NewAIPlugin() *AIPlugin {
	return &AIPlugin{}
}

// ID implements base.ShellPlugin.
func (a *AIPlugin) ID() string {
	return "ai"
}

// Install implements base.ShellPlugin.
func (a *AIPlugin) Install(shell *base.Shell) error {
	c := openai.NewClient()
	a.client = &c

	a.syncLimits()
	return nil
}

// PrepareContext implements base.ShellPlugin.
func (a *AIPlugin) PrepareContext(ce *base.CommandExecution, shell *base.Shell) (context.Context, error) {
	return nil, nil
}

// BeforeExecute implements base.ShellPlugin.
func (a *AIPlugin) BeforeExecute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) error {
	qa := sce.QA()
	qa.Question = strings.TrimSpace(strings.Join(sce.Fields(), " "))
	return nil
}

func isNotFoundError(err error) bool {
	return errors.Is(err, interp.ExitStatus(127))
}

// Execute implements base.ShellPlugin.
func (a *AIPlugin) Execute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) (ok bool, err error) {
	switch sce.Mode() {
	case base.ShellModeAuto:
		if err := sce.DefaultExecHandler(); err != nil {
			if !isNotFoundError(err) {
				return true, err
			}
		} else {
			return true, nil
		}
	case base.ShellModeUser:
		return true, sce.DefaultExecHandler()
	case base.ShellModeAI:
	default:
		return false, nil
	}

	qa := sce.QA()
	iter := 0

	for {
		if iter > a.iterationLimit {
			iter--
			break
		}

		messages, err := a.retrieveMessages(ce, shell, qa)
		if err != nil {
			return true, err
		}

		stream := a.client.Chat.Completions.NewStreaming(
			ce.Context(),
			openai.ChatCompletionNewParams{
				Model:    base.GetConfig(base.ConfigOpenAIModel),
				Messages: messages,
				Tools:    a.retrieveToolDefinitions(),
			},
			option.WithAPIKey(base.GetConfig(base.ConfigOpenAIAPIKey)),
			option.WithBaseURL(base.GetConfig(base.ConfigOpenAIBaseURL)),
		)

		// optionally, an accumulator helper can be used
		acc := openai.ChatCompletionAccumulator{}

		isLeadingSpace := true
		hasToolCall := false
		hasText := false

		if sce.ColorSupported() {
			fmt.Fprint(sce.Stdout(), base.ColorGray)
		}
		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			if len(chunk.Choices) == 0 {
				continue
			}

			text := chunk.Choices[0].Delta.Content
			if isLeadingSpace {
				text = strings.TrimLeftFunc(text, unicode.IsSpace)
				if text == "" {
					continue
				}
			}
			isLeadingSpace = false
			if text != "" {
				hasText = true
				sce.Stdout().Write([]byte(text))
			}
		}
		if sce.ColorSupported() {
			fmt.Fprint(sce.Stdout(), base.ColorReset)
		}

		if err := stream.Err(); err != nil {
			return true, err
		}

		if hasText {
			sce.Stdout().Write([]byte("\n"))
		}

		if len(acc.Choices) > 0 {
			choice := acc.Choices[0]

			if len(choice.Message.ToolCalls) > 0 {
				hasToolCall = true

				// Flush the leading answer text if it exists, and ensure the buffer is clean before handling the tool call
				if answerText := ce.AnswerText(); answerText != "" {
					qa.Answers = append(qa.Answers, base.AIAssistantAnswer{
						Text:     answerText,
						ToolCall: nil,
					})
					ce.Buffer().Reset()
				}

				toolCall := &choice.Message.ToolCalls[0]
				err = a.handleToolCall(ce, sce, toolCall, shell)

				if err != nil {
					if answerText := ce.AnswerText(); answerText != "" {
						qa.Answers = append(qa.Answers, base.AIAssistantAnswer{
							Text:     answerText,
							ToolCall: toolCall,
						})
					} else {
						qa.Answers = append(qa.Answers, base.AIAssistantAnswer{
							Text:     fmt.Sprintf("Error: %s", err.Error()),
							ToolCall: toolCall,
						})
					}
				}
				ce.Buffer().Reset()
			}
		}

		if !hasToolCall {
			break
		}
		iter++
	}
	return true, nil
}

func (a *AIPlugin) AfterExecute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) error {
	if strings.EqualFold(sce.Cmd(), "reset") {
		a.historyExecutions = nil
		return nil
	}

	qa := sce.QA()
	if qa.IsRoot() {
		switch sce.Cmd() {
		case string(ExtensionCommandUserMode):
			fallthrough
		case string(ExtensionCommandUserModeShort):
			fallthrough
		case string(ExtensionCommandAIMode):
			fallthrough
		case string(ExtensionCommandAISet):
			fallthrough
		case string(ExtensionCommandAIGet):
			fallthrough
		case string(ExtensionCommandAIPrompt):
			fallthrough
		case string(ExtensionCommandAITool):
			fallthrough
		case string(ExtensionCommandHistory):
		default:
			defer ce.AppendQA(qa)
		}
	}

	if err := sce.Error(); err != nil {
		if exitStatus, ok := err.(interp.ExitStatus); ok {
			fmt.Fprintln(sce.Stdai(), a.formatExitStatus(exitStatus))
		} else {
			shell.PrintError(sce.Stdai(), err)
		}
	}

	toolCall := qa.UnderToolCall
	trace := qa.Trace()

	if answerText := ce.AnswerText(); answerText != "" {
		for _, qa := range trace {
			qa.Answers = append(qa.Answers, base.AIAssistantAnswer{
				Text:     answerText,
				ToolCall: toolCall,
			})
		}
	} else {
		for _, qa := range trace {
			qa.Answers = append(qa.Answers, a.generateFallbackAssistantAnswer(sce.Error(), toolCall))
		}
	}
	ce.Buffer().Reset()
	return nil
}

// End implements base.ShellPlugin.
func (a *AIPlugin) End(ce *base.CommandExecution, shell *base.Shell) error {
	a.syncLimits()

	if len(ce.QA()) == 0 {
		return nil
	}

	limit := a.historyLimit
	if limit == 0 {
		a.historyExecutions = nil
		return nil
	}
	a.historyExecutions = append(a.historyExecutions, ce.QA()...)

	n := len(a.historyExecutions)
	if n > limit {
		a.historyExecutions = a.historyExecutions[(n - limit):]
	}
	return nil
}

// AutoComplete implements base.ShellPlugin.
func (a *AIPlugin) AutoComplete(line []rune, pos int, shell *base.Shell) (newLine [][]rune, length int) {
	return nil, 0
}

// GeneratePrompt implements base.ShellPlugin.
func (a *AIPlugin) GeneratePrompt(ce *base.CommandExecution, shell *base.Shell) (ok bool, prompt string, err error) {
	return false, "", nil
}

func (a *AIPlugin) evalToolCall(
	code []byte,
	ce *base.CommandExecution,
	sce *base.SubCommandExecution,
	toolCall *openai.ChatCompletionMessageToolCall,
	shell *base.Shell,
) error {
	if len(code) == 0 {
		return nil
	}

	isBuiltin := true
	qa := sce.QA()
	// if the modifier function is never triggered, it means the command is a builtin command
	err := shell.Eval(ce, code, func(child *base.SubCommandExecution) {
		isBuiltin = false
		child.Inherit(sce)
		child.SetMode(base.ShellModeUser)
		child.QA().UnderToolCall = toolCall
	})

	if err != nil {
		return err
	}

	if isBuiltin {
		if answerText := ce.AnswerText(); answerText != "" {
			qa.Answers = append(qa.Answers, base.AIAssistantAnswer{
				Text:     answerText,
				ToolCall: toolCall,
			})
		} else {
			qa.Answers = append(qa.Answers, a.generateFallbackAssistantAnswer(nil, toolCall))
		}
	}
	return err
}

func (a *AIPlugin) handleToolCall(ce *base.CommandExecution, sce *base.SubCommandExecution, toolCall *openai.ChatCompletionMessageToolCall, shell *base.Shell) error {
	toolName := toolCall.Function.Name

	switch {
	case toolName == string(ToolNameExecute):
		params := AIExecToolParams{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return err
		}
		params.Code = strings.TrimSpace(params.Code)
		if params.Code == "" {
			return nil
		}
		if sce.ColorSupported() {
			fmt.Fprintf(sce.Stdout(), "%suse:%s \033[4;34m%s\033[0m\n\n", base.ColorBlue, base.ColorReset, params.Code)
		} else {
			fmt.Fprintf(sce.Stdout(), "use: %s\n\n", params.Code)
		}

		return a.evalToolCall([]byte(params.Code), ce, sce, toolCall, shell)

	case strings.HasPrefix(toolName, string(ToolNameUserDefinedPrefix)):
		toolName := strings.TrimPrefix(toolName, string(ToolNameUserDefinedPrefix))
		tool, ok := base.GetDefinedTool(toolName)
		if !ok {
			return fmt.Errorf("%s: tool not found", toolName)
		}
		params := AIUserToolParams{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
			return err
		}

		stmt, err := base.CombineFields(append([]string{tool.Entrypoint}, params.Args...))
		if err != nil {
			return err
		}

		if sce.ColorSupported() {
			fmt.Fprintf(sce.Stdout(), "%suse tool:%s \033[4;34m%s\033[0m\n\n", base.ColorBlue, base.ColorReset, stmt)
		} else {
			fmt.Fprintf(sce.Stdout(), "use tool: %s\n\n", stmt)
		}

		return a.evalToolCall([]byte(stmt), ce, sce, toolCall, shell)

	default:
		return fmt.Errorf("%s: tool not found", toolCall.Function.Name)
	}
}

func (a *AIPlugin) formatExitStatus(err interp.ExitStatus) string {
	switch err {
	case 130:
		return fmt.Sprintf("Exit status %d: command cancelled by user", err)
	case 131:
		return fmt.Sprintf("Exit status %d: segment fault", err)
	}
	return fmt.Sprintf("Exit status: %d\n", err)
}

func (a *AIPlugin) retrieveMessages(ce *base.CommandExecution, shell *base.Shell, extra *base.AIExecution) ([]openai.ChatCompletionMessageParamUnion, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(a.historyExecutions)*(1+a.iterationLimit)+1)

	if systemPrompt, err := a.generateSystemPrompt(shell); err != nil {
		return nil, err
	} else if systemPrompt != "" {
		messages = append(messages, openai.SystemMessage(systemPrompt))
	}

	appendQA := func(qa *base.AIExecution) {
		messages = append(messages, openai.UserMessage(qa.Question))

		for _, answer := range qa.Answers {
			if answer.ToolCall != nil {
				messages = append(
					messages,
					openai.ChatCompletionMessageParamUnion{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Content: openai.ChatCompletionAssistantMessageParamContentUnion{
								OfString: openai.String(""),
							},
							ToolCalls: []openai.ChatCompletionMessageToolCallParam{
								{
									ID:   answer.ToolCall.ID,
									Type: "function",
									Function: openai.ChatCompletionMessageToolCallFunctionParam{
										Name:      answer.ToolCall.Function.Name,
										Arguments: answer.ToolCall.Function.Arguments,
									},
								},
							},
						},
					},
					openai.ToolMessage(answer.Text, answer.ToolCall.ID),
				)
			} else if answer.Text != "" {
				messages = append(messages, openai.AssistantMessage(answer.Text))
			}
		}
	}

	// append the history executions
	for _, qa := range a.historyExecutions {
		appendQA(qa)
	}

	// append the current AI execution
	for _, qa := range ce.QA() {
		appendQA(qa)
	}

	if extra != nil {
		appendQA(extra)
	}

	return messages, nil
}

func (a *AIPlugin) retrieveToolDefinitions() []openai.ChatCompletionToolParam {
	tools := []openai.ChatCompletionToolParam{
		{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        string(ToolNameExecute),
				Description: openai.String("Execute code in parameter, which means you will do: `source [code]`"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"code": map[string]any{
							"type":        "string",
							"description": "The code to execute",
						},
					},
					"required": []string{"code"},
				},
			},
		},
	}

	definedTools := base.GetDefinedTools()
	for _, tool := range definedTools {
		usage := "none"
		if tool.Usage != "" {
			usage = tool.Usage
		}

		tools = append(tools, openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name: string(ToolNameUserDefinedPrefix) + tool.Name,
				Description: openai.String(
					fmt.Sprintf("Execute %s, usage: %s", tool.Entrypoint, usage),
				),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"args": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "string",
							},
						},
					},
				},
			},
		})
	}

	return tools
}

func (a *AIPlugin) generateFallbackAssistantAnswer(err error, toolCall *openai.ChatCompletionMessageToolCall) base.AIAssistantAnswer {
	if err != nil {
		return base.AIAssistantAnswer{
			Text:     fmt.Sprintf("Error: %s", err.Error()),
			ToolCall: toolCall,
		}
	} else {
		return base.AIAssistantAnswer{
			Text:     "done",
			ToolCall: toolCall,
		}
	}
}

func (a *AIPlugin) generateSystemPrompt(shell *base.Shell) (string, error) {
	state := shell.State()
	sb := strings.Builder{}

	cmd := base.DefaultFileName
	if cmd != shell.FileName() {
		cmd = fmt.Sprintf("%s %s", cmd, shell.FileName())
	}

	if err := systemPromptTemplate.Execute(&sb, map[string]any{
		"prompt": base.GetDefinedSystemPrompts(),
		"cmd":    cmd,
		"shell":  base.DefaultFileName,
		"os":     state.OS(),
		"arch":   state.Arch(),
		"wd":     shell.Dir(),
		"user":   state.User().Username,
	}); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (a *AIPlugin) syncLimits() {
	if limit, ok := base.GetIntConfig(base.ConfigMaxHistory); ok {
		a.historyLimit = max(limit, 0)
	}
	if limit, ok := base.GetIntConfig(base.ConfigMaxIterations); ok {
		a.iterationLimit = max(limit, 1)
	}
}
