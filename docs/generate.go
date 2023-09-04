package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra/doc"
	"github.com/walteh/retab/cmd/root"
	"github.com/walteh/snake"
)

func run(ctx context.Context, ref string) error {
	log.SetFlags(0)

	cmd := snake.NewRootCommand(ctx, &root.Root{})

	cmd.DisableAutoGenTag = true

	mdpath := filepath.Join(ref, "md")

	if err := os.MkdirAll(mdpath, 0755); err != nil {
		return err
	}

	err := doc.GenMarkdownTree(cmd, mdpath)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	ctx := context.Background()

	ref := "./docs/reference/"
	if len(os.Args) > 1 {
		ref = os.Args[1]
	}

	if err := run(ctx, ref); err != nil {
		log.Printf("ERROR: %+v", err)
		os.Exit(1)
	}
}
