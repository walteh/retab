package root

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type args struct {
	data     string
	filename string
}

type test struct {
	name    string
	args    args
	want    string
	wantErr bool
}

var tests = []test{
	{
		name: "test",
		args: args{
			data: `
resource "aws_s3_bucket" "b" {
bucket = "my-tf-test-bucket"
acl = "private"
}`,
			filename: "test.tf",
		},
		want: `
resource "aws_s3_bucket" "b" {
	bucket = "my-tf-test-bucket"
	acl    = "private"
}`,
	},
}

func (tr *test) run(ctx context.Context, t *testing.T, runner func(ctx context.Context, strs ...string) error) {
	t.Helper()

	d := os.TempDir()

	// save sample 1 to a temp file
	f, err := os.CreateTemp(d, tr.args.filename)
	if err != nil {
		t.Errorf("CreateTemp() error = %v", err)
		return
	}

	// 	c, err :=

	// 	// make tmp editorconfig
	// 	def := `
	// root = true

	// [*]
	// indent_style = tabs
	// indent_size = 4
	// trim_trailing_whitespace = true
	// trim_multiple_empty_lines = true
	// `

	// 	// write editorconfig
	// 	err = os.WriteFile(".editorconfig", []byte(def), 0644)

	_, err = f.WriteString(tr.args.data)
	if err != nil {
		t.Errorf("WriteString() error = %v", err)
		return
	}

	err = f.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
		return
	}

	t.Cleanup(func() {
		os.Remove(f.Name())
	})

	err = runner(ctx, "retab", "hcl", "--file", f.Name(), "--debug")
	if err != nil {
		t.Errorf("runner() error = %v", err)
		return
	}

	// read the file back
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Errorf("ReadFile() error = %v", err)
		return
	}

	assert.Equal(t, tr.want, string(b))
}

func TestRootUnit(t *testing.T) {

	ctx := context.Background()

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			runner := func(ctx context.Context, strs ...string) error {
				_, got, err := NewCommand(ctx)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewCommand() error = %v, wantErr %v", err, tt.wantErr)
					return err
				}

				os.Args = strs
				err = got.ExecuteContext(ctx)
				if err != nil {
					t.Errorf("ExecuteContext() error = %v", err)
					return err
				}

				return nil
			}

			tt.run(ctx, t, runner)

		})
	}
}

func TestRootE2E(t *testing.T) {

	if os.Getenv("E2E") != "1" {
		t.SkipNow()
	}
	ctx := context.Background()

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			runner := func(ctx context.Context, strs ...string) error {
				cmd := exec.CommandContext(ctx, strs[0], strs[1:]...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return cmd.Run()
			}

			tt.run(ctx, t, runner)

		})
	}

}
