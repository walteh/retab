package install

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/buildrc/pkg/install"
	"github.com/walteh/snake"
)

var _ snake.Snakeable = (*Handler)(nil)

type Handler struct {
	Latest bool
}

func (me *Handler) BuildCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Short: "install buildrc",
	}

	cmd.Args = cobra.ExactArgs(0)

	cmd.PersistentFlags().BoolVarP(&me.Latest, "latest", "l", false, "Install the latest version")

	return cmd
}

func (me *Handler) ParseArguments(ctx context.Context, cmd *cobra.Command, file []string) error {

	return nil

}

func (me *Handler) Run(ctx context.Context, cmd *cobra.Command) error {
	if me.Latest {
		return install.InstallLatestGithubRelease(ctx, afero.NewOsFs(), afero.NewOsFs(), "walteh", "retab", "")
	}
	return install.InstallSelfAs(ctx, afero.NewOsFs(), "retab")
}
