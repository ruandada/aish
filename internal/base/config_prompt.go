package base

import "strings"

var definedSystemPrompts strings.Builder

const defaultDefinedSystemPrompts = "You are a smart assistant running on a UNIX-like shell."

func AddDefinedSystemPrompt(prompt string) {
	definedSystemPrompts.WriteString(prompt + "\n\n")
}

func ClearDefinedSystemPrompts() {
	definedSystemPrompts.Reset()
}

func GetDefinedSystemPrompts() string {
	if s := definedSystemPrompts.String(); s != "" {
		return s
	}

	return defaultDefinedSystemPrompts
}
