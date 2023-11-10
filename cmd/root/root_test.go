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

func (tr *test) run(t *testing.T, ctx context.Context, runner func(ctx context.Context, strs ...string) error) {
	t.Helper()

	// save sample 1 to a temp file
	f, err := os.CreateTemp("", tr.args.filename)
	if err != nil {
		t.Errorf("NewCommand() error = %v", err)
		return
	}

	_, err = f.WriteString(tr.args.data)
	if err != nil {
		t.Errorf("NewCommand() error = %v", err)
		return
	}

	t.Cleanup(func() {
		os.Remove(f.Name())
	})

	err = runner(ctx, "retab", "hcl", "--file", f.Name(), "--debug")
	if err != nil {
		t.Errorf("NewCommand() error = %v", err)
		return
	}

	// read the file back
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Errorf("NewCommand() error = %v", err)
		return
	}

	assert.Equal(t, tr.want, string(b))
}

func TestRootUnit(t *testing.T) {

	ctx := context.Background()

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			runner := func(ctx context.Context, strs ...string) error {
				got, err := NewCommand()
				if (err != nil) != tt.wantErr {
					t.Errorf("NewCommand() error = %v, wantErr %v", err, tt.wantErr)
					return err
				}

				os.Args = strs
				err = got.ExecuteContext(ctx)
				if err != nil {
					t.Errorf("NewCommand() error = %v", err)
					return err
				}

				return nil
			}

			tt.run(t, ctx, runner)

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
				cmd := exec.Command(strs[0], strs[1:]...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return cmd.Run()
			}

			tt.run(t, ctx, runner)

		})
	}

}
