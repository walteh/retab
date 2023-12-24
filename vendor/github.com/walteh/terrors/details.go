package terrors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/rs/zerolog"
)

type stringWriter struct {
	*strings.Builder
}

func (s *stringWriter) Write(b []byte) (int, error) {
	str, err := FormatJsonForDetail(b, []string{"level", "error", "caller"}, []string{"package", "file", "message", "function", "chain"})
	if err != nil {
		return 0, err
	}

	return s.Builder.WriteString(str)
}

func (e *wrapError) Detail() string {
	srtwrite := &strings.Builder{}
	w1 := zerolog.New(&stringWriter{srtwrite})
	pkg, funct, filestr, linestr := e.Frame().Location()

	ed := w1.Err(nil).
		Str("package", pkg).
		Str("file", fmt.Sprintf("%s:%d", filestr, linestr)).
		Str("message", e.msg).
		Str("function", funct)

	if e.err != nil {
		ed = ed.AnErr("chain", e.err)
	}

	for _, ev := range e.event {
		ed = ev(ed)
	}

	ed.Send()

	return srtwrite.String()
}

func FormatJsonForDetail(b []byte, ignored []string, priority []string) (string, error) {
	dat := map[string]any{}

	err := json.Unmarshal(b, &dat)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)

	wrt := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)

	wrtfunc := func(t string, k string, v any, newline bool) error {

		out := fmt.Sprintf("%v", v)

		if v == nil {
			out = "nil"
		}

		if out == "" {
			out = "\"\"" // empty string
		}

		if newline {
			out = out + "\n"
		}

		_, err := wrt.Write([]byte(fmt.Sprintf("%s%s\t= %s", t, k, out)))
		return err
	}

	priorityKeys := []string{}
	normal := []string{}
	for k := range dat {
		if slices.Contains(ignored, k) {
			continue
		}
		if slices.Contains(priority, k) {
			priorityKeys = append(priorityKeys, k)
		} else {
			normal = append(normal, k)
		}
	}

	slices.Sort(priorityKeys)
	slices.Sort(normal)

	all := append(priorityKeys, normal...)

	for i, k := range all {
		if err = wrtfunc("", k, dat[k], i != len(all)-1); err != nil {
			return "", err
		}
	}

	err = wrt.Flush()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
