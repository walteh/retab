package main

import (
	"context"
	"log"

	"github.com/spf13/afero"
	"github.com/walteh/tftab/cmd/root"
	"github.com/walteh/tftab/pkg/cli"
)

func main() {

	ctx := context.Background()

	rootCmd := cli.RegisterRoot(ctx, &root.Root{})

	ctx = cli.Bind(ctx, (*afero.Fs)(nil), afero.NewOsFs())

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatalf("ERROR: %+v", err)
	}

}
