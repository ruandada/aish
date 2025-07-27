package base

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func (s *Shell) PrintError(writer io.Writer, err error) {
	if writer == nil {
		return
	}
	fmt.Fprintf(writer, "%s: %v\n", s.fileName, err)
}

func ReaderDescriptor(reader io.Reader) (*os.File, error) {
	if f, ok := reader.(*os.File); ok {
		return f, nil
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	go func() {
		io.Copy(pw, reader)
		pw.Close()
	}()

	return pr, nil
}

func WriterDescriptor(writer io.Writer) (*os.File, error) {
	if f, ok := writer.(*os.File); ok {
		return f, nil
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	go func() {
		io.Copy(writer, pr)
		pr.Close()
	}()

	return pw, nil
}

func CombineFields(fields []string) (string, error) {
	sb := strings.Builder{}
	encoder := json.NewEncoder(&sb)
	encoder.SetEscapeHTML(false)

	n := len(fields)

	for i, field := range fields {
		if err := encoder.Encode(field); err != nil {
			return "", err
		}
		if i < n-1 {
			sb.WriteString(" ")
		}
	}
	return strings.ReplaceAll(sb.String(), "\n", ""), nil
}

func LookPath(file string, wd string, PATH string) (string, error) {
	if filepath.IsAbs(file) {
		return filepath.Clean(file), nil
	}

	abs := func(file string, wd string) string {
		if file == "" {
			file = "."
		}

		if filepath.IsAbs(file) {
			return filepath.Clean(file)
		}

		return filepath.Clean(filepath.Join(wd, file))
	}

	dirs := append([]string{wd}, filepath.SplitList(PATH)...)
	for _, dir := range dirs {
		path := filepath.Join(abs(dir, wd), file)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		return filepath.Clean(path), nil
	}

	return "", os.ErrNotExist
}

func FindExecutableNames(PATH string, wd string) ([]string, error) {
	var execs map[string]bool = map[string]bool{}
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	for _, dir := range filepath.SplitList(PATH) {
		if !filepath.IsAbs(dir) {
			continue
		}

		wg.Add(1)
		go func(dir string) {
			defer wg.Done()

			files, err := os.ReadDir(dir)
			if err != nil {
				return
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}

				mu.Lock()
				execs[file.Name()] = true
				mu.Unlock()
			}
		}(dir)
	}

	wg.Wait()
	names := []string{}
	for name := range execs {
		names = append(names, name)
	}
	return names, nil
}
