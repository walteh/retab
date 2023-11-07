package root

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

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
			os.Args = []string{"retab", "hcl", "../../retab.hcl"}
			err = got.ExecuteContext(tt.args.ctx)
			if err != nil {
				t.Errorf("NewCommand() error = %v", err)
				return
			}
		})
	}
}
