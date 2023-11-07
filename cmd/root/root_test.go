package root

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

const sample1 = `
BRANCH = "main"

func "leggo" {
	params = [abc, def]
	result = abc + def
}

file "default.yaml" {
	dir    = "./.github/workflows"
	schema = "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json"
	data = {
		name = "test"

		on = {
			push = {
				branches = [BRANCH]
			}
		}
		jobs = {
			build = {
				runs-on = "ubuntu-latest"
				steps = [
					{
						name = "Checkout"
						uses = "actions/checkout@v2"
						with = {
							fetch-depth = leggo(1, 2)
						}
					},
					{
						name = "Run tests"
						run  = <<SHELL
							echo "Hello world"
						SHELL
					},
				]

			}
		}
	}
}
`

func TestNewCommand(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *cobra.Command
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx: context.Background(),
			},
			want: &cobra.Command{
				Use:   "retab",
				Short: "retab brings tabs to terraform",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCommand()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// save sample 1 to a temp file
			f, err := os.CreateTemp("", "retab.hcl")
			if err != nil {
				t.Errorf("NewCommand() error = %v", err)
				return
			}

			_, err = f.WriteString(sample1)
			if err != nil {
				t.Errorf("NewCommand() error = %v", err)
				return
			}

			t.Cleanup(func() {
				os.Remove(f.Name())
			})

			os.Args = []string{"retab", "hcl", f.Name()}
			err = got.ExecuteContext(tt.args.ctx)
			if err != nil {
				t.Errorf("NewCommand() error = %v", err)
				return
			}
		})
	}
}
