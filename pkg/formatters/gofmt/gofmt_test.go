package gofmt_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/walteh/retab/v2/gen/mocks/pkg/formatmock"
	"github.com/walteh/retab/v2/pkg/diff"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/formatters/gofmt"
)

func formatGo(ctx context.Context, cfg format.Configuration, src []byte) (string, error) {
	formatter := gofmt.NewFormatter()
	reader, err := formatter.Format(ctx, cfg, bytes.NewReader(src))
	if err != nil {
		return "", err
	}

	result, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func TestGo(t *testing.T) {
	start := `package main
import "fmt"
func  main ( ) {
fmt.Println("hello world")
}
`

	expected := `package main

import "fmt"

func main() {
    fmt.Println("hello world")
}
`

	cfg := formatmock.NewMockConfiguration(t)
	cfg.EXPECT().UseTabs().Return(false).Maybe()
	cfg.EXPECT().IndentSize().Return(4).Maybe()
	cfg.EXPECT().Raw().Return(map[string]string{
		"go_module_name":       "github.com/walteh/retab/v2",
		"go_yes_i_want_spaces": "true",
	}).Maybe()

	actual, err := formatGo(t.Context(), cfg, []byte(start))
	if err != nil {
		t.Fatal(err)
	}

	diff.Require(t).Want(expected).Got(actual).Equals()
}

func TestGoImports(t *testing.T) {
	start := `package main

import (
	"github.com/walteh/retab/v2/pkg/format" // comment abc
	"fmt"
	myAlias "fun/util" // comment def
	"github.com/gin-gonic/gin"

	. "github.com/onsi/gomega" // dot import
	"golang.org/x/sync/errgroup"
)

func main() {
	fmt.Println("hello")
	gin.New()
	format.NewFormatter()
	Expect(true).To(BeTrue())
	myAlias.DoSomething()
}
`

	expected := `package main

import (
	. "github.com/onsi/gomega" // dot import

	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"

	myAlias "more-fun/util" // comment def

	"github.com/walteh/retab/v2/pkg/format" // comment abc
)

func main() {
	fmt.Println("hello")
	gin.New()
	format.NewFormatter()
	Expect(true).To(BeTrue())
	myAlias.DoSomething()
}
`

	cfg := formatmock.NewMockConfiguration(t)
	cfg.EXPECT().UseTabs().Return(true).Maybe() // gofmt.Source doesn't use this, but goimports-reviser might
	cfg.EXPECT().IndentSize().Return(4).Maybe() // gofmt.Source doesn't use this, but goimports-reviser might
	cfg.EXPECT().Raw().Return(map[string]string{
		"go_module_name":    "github.com/walteh/retab/v2",
		"go_rename_imports": "fun/util=more-fun/util",
		// "rename_imports_separator": "=", // Assuming default is '='
	}).Maybe()

	actual, err := formatGo(t.Context(), cfg, []byte(start))
	if err != nil {
		t.Fatal(err)
	}

	diff.Require(t).Want(expected).Got(actual).Equals()
}
