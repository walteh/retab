package yamlwrite2

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/braydonk/yaml"
	"github.com/google/yamlfmt"
	"github.com/google/yamlfmt/formatters/basic"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
)

const lineBreakPlaceholder = "#magic___^_^___line"

type paddinger struct {
	strings.Builder
}

func (p *paddinger) adjust(txt string) {
	var indentSize int
	for i := 0; i < len(txt) && txt[i] == ' '; i++ { // yaml only allows space to indent.
		indentSize++
	}
	// Grows if the given size is larger than us and always return the max padding.
	for diff := indentSize - p.Len(); diff > 0; diff-- {
		p.WriteByte('	')
	}
}

func MakeFeatureRetainLineBreak(linebreakStr string) yamlfmt.Feature {
	return yamlfmt.Feature{
		Name:         "Retain Line Breaks",
		BeforeAction: replaceLineBreakFeature(linebreakStr),
		AfterAction:  restoreLineBreakFeature(linebreakStr),
	}
}

func replaceLineBreakFeature(newlineStr string) yamlfmt.FeatureFunc {
	return func(content []byte) ([]byte, error) {
		var buf bytes.Buffer
		reader := bytes.NewReader(content)
		scanner := bufio.NewScanner(reader)
		var padding paddinger
		for scanner.Scan() {
			txt := scanner.Text()
			padding.adjust(txt)
			if strings.TrimSpace(txt) == "" { // line break or empty space line.
				buf.WriteString(padding.String()) // prepend some padding incase literal multiline strings.
				buf.WriteString(lineBreakPlaceholder)
				buf.WriteString(newlineStr)
				continue
			}
			buf.WriteString(txt)
			buf.WriteString(newlineStr)
		}
		return buf.Bytes(), scanner.Err()
	}
}

func restoreLineBreakFeature(newlineStr string) yamlfmt.FeatureFunc {
	return func(content []byte) ([]byte, error) {
		var buf bytes.Buffer
		reader := bytes.NewReader(content)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			txt := scanner.Text()
			if strings.TrimSpace(txt) == "" {
				// The basic yaml lib inserts newline when there is a comment(either placeholder or by user)
				// followed by optional line breaks and a `---` multi-documents.
				// To fix it, the empty line could only be inserted by us.
				continue
			}
			if strings.HasPrefix(strings.TrimLeft(txt, " "), lineBreakPlaceholder) {
				buf.WriteString(newlineStr)
				continue
			}
			buf.WriteString(txt)
			buf.WriteString(newlineStr)
		}
		return buf.Bytes(), scanner.Err()
	}
}

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Targets() []string {
	return []string{"*.yaml", "*.yml"}
}

type FlowInterface struct {
}

func (me *Formatter) Format(_ context.Context, cfg configuration.Provider, read io.Reader) (io.Reader, error) {
	f := newFormatter(basic.DefaultConfig())

	all, err := io.ReadAll(read)
	if err != nil {
		return nil, err
	}

	// dec := yaml.NewDecoder(read)

	// v := interface{}(nil)
	// if err := dec.Decode(&v); err != nil {
	// 	return nil, err
	// }

	feat := yamlfmt.Feature{
		Name:         "Retain Line Breaks",
		BeforeAction: replaceLineBreakFeature("\n"),
		AfterAction:  restoreLineBreakFeature("\n"),
	}

	f.Features = append(f.Features, feat)

	f.YAMLFeatures = append(f.YAMLFeatures, func(node yaml.Node) error {
		node.Style = yaml.FlowStyle
		return nil
	})

	s, err := f.Format(all)
	if err != nil {
		return nil, err
	}

	return strings.NewReader(string(s)), nil
}

func newFormatter(config *basic.Config) *basic.BasicFormatter {
	return &basic.BasicFormatter{
		Config:   config,
		Features: basic.ConfigureFeaturesFromConfig(config),
	}
}

// func TestFormatterRetainsComments(t *testing.T) {
// 	f := newFormatter(basic.DefaultConfig())

// 	yaml := `x: "y" # foo comment`

// 	s, err := f.Format([]byte(yaml))
// 	if err != nil {
// 		t.Fatalf("expected formatting to pass, returned error: %v", err)
// 	}
// 	if !strings.Contains(string(s), "#") {
// 		t.Fatal("comment was stripped away")
// 	}
// }

// func TestFormatterPreservesKeyOrder(t *testing.T) {

// 	unmarshalledStr := string(s)
// 	bPos := strings.Index(unmarshalledStr, "b")
// 	aPos := strings.Index(unmarshalledStr, "a")
// 	if bPos > aPos {
// 		t.Fatalf("keys were reordered:\n%s", s)
// 	}
// }
