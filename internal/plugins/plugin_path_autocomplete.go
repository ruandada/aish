package plugins

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ruandada/aish/internal/base"
)

type PathAutocompletePlugin struct {
	fragmentCompleter *base.FragmentCompleter
}

var _ base.ShellPlugin = (*PathAutocompletePlugin)(nil)

func NewPathAutocompletePlugin() *PathAutocompletePlugin {
	return &PathAutocompletePlugin{}
}

// Install implements base.ShellPlugin.
func (p *PathAutocompletePlugin) Install(shell *base.Shell) error {
	p.fragmentCompleter = base.NewFragmentCompleter(func(lastFragment []rune, rest []rune) (completions [][]rune) {
		state := shell.State()

		dir, fileNamePrefix := filepath.Split(string(lastFragment))
		if dir == "" {
			if fileNamePrefix == "." {
				dir, fileNamePrefix = ".", "."
			} else if fileNamePrefix == ".." || strings.HasPrefix(fileNamePrefix, "~") {
				dir, fileNamePrefix = fileNamePrefix, ""
			} else {
				dir = "."
			}
		}

		home := state.User().HomeDir
		if dir == "~" {
			completions = append(completions, []rune(
				"~/",
			))

			// TODO: get all unix users
			completions = append(completions, []rune(
				"~"+state.User().Username,
			))
			return completions
		}

		var lookupDir string
		switch {
		case dir == ".":
			lookupDir = shell.Dir()
		case strings.HasPrefix(dir, "~"):
			restPath := strings.Join(strings.Split(dir, string(os.PathSeparator))[1:], string(os.PathSeparator))
			lookupDir = filepath.Join(home, restPath)
		case filepath.IsAbs(dir):
			lookupDir = dir
		default:
			lookupDir = filepath.Join(shell.Dir(), dir)
		}

		// Read the directory contents
		entries, err := os.ReadDir(lookupDir)
		if err != nil {
			return nil
		}

		for _, entry := range entries {
			name := entry.Name()

			// Skip hidden files, unless the file name prefix starts with .
			if strings.HasPrefix(name, ".") && !strings.HasPrefix(fileNamePrefix, ".") {
				continue
			}

			// Check if the file name matches the prefix
			if strings.HasPrefix(strings.ToLower(name), strings.ToLower(fileNamePrefix)) {
				completionName := name

				// If it's a directory, add a path separator
				if entry.IsDir() {
					completionName += string(os.PathSeparator)
				} else if entry.Type() == os.ModeSymlink {
					func() {
						link, err := os.Readlink(filepath.Join(lookupDir, name))
						if err != nil {
							return
						}

						if !filepath.IsAbs(link) {
							link = filepath.Join(lookupDir, link)
						}
						stat, err := os.Stat(link)

						if err != nil {
							return
						}

						if stat.IsDir() {
							completionName += string(os.PathSeparator)
						}
					}()
				}
				if dir == "." {
					completions = append(completions, []rune(completionName))
				} else if strings.HasSuffix(dir, string(os.PathSeparator)) {
					completions = append(completions, []rune(dir+completionName))
				} else {
					completions = append(completions, []rune(dir+string(os.PathSeparator)+completionName))
				}
			}
		}
		return completions
	})
	return nil
}

// AutoComplete implements base.ShellPlugin.
func (p *PathAutocompletePlugin) AutoComplete(line []rune, pos int, shell *base.Shell) (newLine [][]rune, length int) {
	return p.fragmentCompleter.Do(line, pos)
}

// AfterExecute implements base.ShellPlugin.
func (p *PathAutocompletePlugin) AfterExecute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) error {
	return nil
}

func (p *PathAutocompletePlugin) BeforeExecute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) error {
	return nil
}

// Execute implements base.ShellPlugin.
func (p *PathAutocompletePlugin) Execute(ce *base.CommandExecution, sce *base.SubCommandExecution, shell *base.Shell) (ok bool, err error) {
	return false, nil
}

// GeneratePrompt implements base.ShellPlugin.
func (p *PathAutocompletePlugin) GeneratePrompt(ce *base.CommandExecution, shell *base.Shell) (ok bool, prompt string, err error) {
	return false, "", nil
}

// ID implements base.ShellPlugin.
func (p *PathAutocompletePlugin) ID() string {
	return "path-autocomplete"
}

// PrepareContext implements base.ShellPlugin.
func (p *PathAutocompletePlugin) PrepareContext(ce *base.CommandExecution, shell *base.Shell) (context.Context, error) {
	return nil, nil
}

func (p *PathAutocompletePlugin) End(ce *base.CommandExecution, shell *base.Shell) error {
	return nil
}
