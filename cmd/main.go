package main

import (
	"context"
	"log"

	"github.com/spf13/afero"
	"github.com/walteh/snake"
	"github.com/walteh/tftab/cmd/root"
)

func main() {

	ctx := context.Background()

	rootCmd := snake.NewRootCommand(ctx, &root.Root{})

	ctx = snake.Bind(ctx, (*afero.Fs)(nil), afero.NewOsFs())

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatalf("ERROR: %+v", err)
	}

}
