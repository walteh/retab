package dartwrite

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"

	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
)

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewBufFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Targets() []string {
	return []string{"*.dart"}
}

func (me *Formatter) Format(_ context.Context, cfg configuration.Provider, read io.Reader) (io.Reader, error) {

	// dart format --output show --summary none

	cmd := exec.Command("dart", "format", "--output", "show", "--summary", "none", "--fix")
	cmd.Stdin = read

	read2, write := io.Pipe()

	cmd.Stdout = write

	go func() {
		if err := cmd.Run(); err != nil {
			err := write.CloseWithError(err)
			if err != nil {
				panic(err)
			}
			return
		}
		if err := write.Close(); err != nil {
			panic(err)
		}
	}()

	// // check if dart exists to help with debugging
	// // we do this here so we don't slow down the formatting process
	// if _, err := exec.LookPath("dart"); err != nil {
	// 	return nil, err
	// }

	read3, err := formatDartFile(cfg, read2)
	if err != nil {
		return nil, err
	}

	return read3, nil
}

// formatDartFile takes a formatted dart file and reformats it
// formatDartFile takes a formatted dart file and reformats it according to the configuration.
func formatDartFile(cfg configuration.Provider, input io.Reader) (io.Reader, error) {
	var output bytes.Buffer
	scanner := bufio.NewScanner(input)
	indentation := "\t"
	if !cfg.UseTabs() {
		indentation = strings.Repeat(" ", cfg.IndentSize())
	}

	previousLineWasEmpty := false
	for scanner.Scan() {
		line := scanner.Text()

		// Apply indentation preference.
		line = strings.Replace(line, "  ", indentation, -1)

		// Trim multiple empty lines if configured.
		if cfg.TrimMultipleEmptyLines() {
			if strings.TrimSpace(line) == "" {
				if previousLineWasEmpty {
					continue
				}
				previousLineWasEmpty = true
			} else {
				previousLineWasEmpty = false
			}
		}

		// Write the modified line to the output buffer.
		_, err := output.WriteString(line + "\n")
		if err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !true {

	}

	return &output, nil
}
