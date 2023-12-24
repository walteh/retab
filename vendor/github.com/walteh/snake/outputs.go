package snake

import (
	"context"
	"encoding/json"
	"io"

	"github.com/walteh/terrors"
)

type Stdin interface {
	io.Reader
}

type Stdout interface {
	io.Writer
}

type Stderr interface {
	io.Writer
}

type OutputHandler interface {
	Stdout() io.Writer
	Stderr() io.Writer
	Stdin() io.Reader
	HandleLongRunningOutput(ctx context.Context, cd Chan, out *LongRunningOutput) error
	HandleRawTextOutput(ctx context.Context, cd Chan, out *RawTextOutput) error
	HandleTableOutput(ctx context.Context, cd Chan, out *TableOutput) error
	HandleJSONOutput(ctx context.Context, cd Chan, out *JSONOutput) error
	HandleNilOutput(ctx context.Context, cd Chan, out *NilOutput) error
	HandleFileOutput(ctx context.Context, cd Chan, out *FileOutput) error
}

func (*LongRunningOutput) IsOutput() {}
func (*RawTextOutput) IsOutput()     {}
func (*TableOutput) IsOutput()       {}
func (*JSONOutput) IsOutput()        {}
func (*NilOutput) IsOutput()         {}
func (*FileOutput) IsOutput()        {}

type Output interface {
	IsOutput()
}

type FileOutput struct {
	Dir   string
	Mkdir bool
	Data  map[string]io.Reader
}

type LongRunningOutput struct {
	Start func(context.Context) error
}

type RawTextOutput struct {
	Data string
}

type TableOutput struct {
	ColumnNames    []string
	RowValueData   [][]any
	RowValueColors [][]string
	RawData        any
}

type JSONOutput struct {
	Data json.RawMessage
}

type NilOutput struct{}

func HandleOutput(ctx context.Context, handler OutputHandler, out Output, cd Chan) error {
	if handler == nil {
		return terrors.Errorf("trying to handle output with no handler provided - %T", out)
	}
	switch t := out.(type) {
	case *LongRunningOutput:
		return handler.HandleLongRunningOutput(ctx, cd, t)
	case *RawTextOutput:
		return handler.HandleRawTextOutput(ctx, cd, t)
	case *TableOutput:
		clength := len(t.ColumnNames)
		if len(t.RowValueData) != len(t.RowValueColors) {
			return terrors.Errorf("table output data (%d) does not match colors (%d)", len(t.RowValueData), len(t.RowValueColors))
		}
		for _, row := range t.RowValueData {
			if len(row) != clength {
				return terrors.Errorf("table output column names (%d) do not match data (%d)", clength, len(row))
			}
		}
		for _, row := range t.RowValueColors {
			if len(row) != clength {
				return terrors.Errorf("table output column names (%d) do not match data (%d)", clength, len(row))
			}
		}
		return handler.HandleTableOutput(ctx, cd, t)
	case *JSONOutput:
		return handler.HandleJSONOutput(ctx, cd, t)
	case *NilOutput:
		return handler.HandleNilOutput(ctx, cd, t)
	case *FileOutput:
		return handler.HandleFileOutput(ctx, cd, t)
	default:
		return terrors.Errorf("unknown output type %T", t)
	}
}
