package base

import (
	"strings"
	"unicode"
)

type FragmentCompleterFunc func(lastFragment []rune, rest []rune) (completions [][]rune)

type FragmentCompleter struct {
	do FragmentCompleterFunc
}

func NewFragmentCompleter(do FragmentCompleterFunc) *FragmentCompleter {
	return &FragmentCompleter{
		do: do,
	}
}

func (c *FragmentCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	l := line[:pos]

	fragPos := strings.LastIndexFunc(string(l), unicode.IsSpace) + 1
	frag := l[fragPos:]
	rest := l[:fragPos]
	offset = len(l)

	newLine = make([][]rune, 0)
	if c := c.do(frag, rest); c != nil {
		for _, suggestion := range c {
			if strings.HasPrefix(string(suggestion), string(frag)) {
				newLine = append(newLine, suggestion[len(frag):])
			}
		}
	}
	return newLine, offset
}
