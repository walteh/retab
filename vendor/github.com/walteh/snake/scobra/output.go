package scobra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/k0kubun/colorstring"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/walteh/snake"
)

// YO could we have more inputs for the ouputs? maybe add some specific flags to them if they have a run method?
// like being able to pass a --json flag to the table output and it will convert it to json? or a csv flag, or a file flag.

var _ snake.OutputHandler = (*OutputHandler)(nil)

func (me *OutputHandler) Stderr() io.Writer {
	return me.cmd.ErrOrStderr()
}

func (me *OutputHandler) Stdout() io.Writer {
	return me.cmd.OutOrStdout()

}

func (me *OutputHandler) Stdin() io.Reader {
	return os.Stdin
}

type OutputHandler struct {
	cmd *cobra.Command
}

func NewOutputHandler(cmd *cobra.Command) *OutputHandler {
	return &OutputHandler{
		cmd: cmd,
	}
}

func (me *OutputHandler) HandleJSONOutput(ctx context.Context, cd snake.Chan, out *snake.JSONOutput) error {
	// Convert the output data to JSON format
	jsonData, err := json.MarshalIndent(out.Data, "", "\t")
	if err != nil {
		return err // Handle or return the error appropriately
	}

	// Print the formatted JSON to the command's output
	me.cmd.Println(string(jsonData))

	return nil
}

// HandleLongRunningOutput implements sbind.OutputHandler.
func (*OutputHandler) HandleLongRunningOutput(ctx context.Context, cd snake.Chan, out *snake.LongRunningOutput) error {
	return out.Start(ctx)
}

// HandleRawTextOutput implements sbind.OutputHandler.
func (me *OutputHandler) HandleRawTextOutput(ctx context.Context, cd snake.Chan, out *snake.RawTextOutput) error {
	me.cmd.Println("")

	me.cmd.Println(out.Data)

	me.cmd.Println("")
	return nil
}

// HandleTableOutput implements sbind.OutputHandler.
func (me *OutputHandler) HandleTableOutput(ctx context.Context, cd snake.Chan, out *snake.TableOutput) error {
	table := tablewriter.NewWriter(me.cmd.OutOrStdout())

	table.SetHeader(out.ColumnNames)

	for i, row := range out.RowValueData {

		strdat := make([]string, len(row))
		for j, v := range row {
			cols := strings.Split(out.RowValueColors[i][j], " ")

			colstr := "[" + strings.Join(cols, ",") + "]"

			if reflect.TypeOf(v).Kind() == reflect.Ptr {
				v = reflect.ValueOf(v).Elem().Interface()
			}
			if v == nil {
				strdat[j] = colorstring.Color(colstr + "nil")
				continue
			}
			strdat[j] = colorstring.Color(fmt.Sprintf("%s%v", colstr, v))
		}

		table.Append(strdat)
	}

	table.Render()

	return nil
}

// HandleNilOutput implements sbind.OutputHandler.
func (me *OutputHandler) HandleNilOutput(ctx context.Context, cd snake.Chan, out *snake.NilOutput) error {
	me.cmd.Println("nil output")
	return nil
}

// HandleFileOutput implements sbind.OutputHandler.
func (me *OutputHandler) HandleFileOutput(ctx context.Context, cd snake.Chan, out *snake.FileOutput) error {
	dir := out.Dir

	if dir == "" {
		dir = "."
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if out.Mkdir {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	me.cmd.Println("")

	me.cmd.Printf("writing %d files to %s\n", len(out.Data), dir)

	for name, content := range out.Data {
		dat, err := io.ReadAll(content)
		if err != nil {
			return err
		}
		me.cmd.Printf("writing %d bytes to %s...", len(dat), name)
		err = os.WriteFile(filepath.Join(dir, name), dat, 0644)
		if err != nil {
			me.cmd.Println("...failed")
			return err
		}
		me.cmd.Println("...done")
	}

	me.cmd.Println("done writing files")

	me.cmd.Println("")

	return nil
}
