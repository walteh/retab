package main

import (
	"context"
	"log"
	"os"

	"github.com/spf13/cobra/doc"
	"github.com/walteh/tftab/cmd/root"
	"github.com/walteh/tftab/pkg/cli"
)

func run() error {

	cmd := cli.RegisterRoot(context.Background(), &root.Root{})

	err := doc.GenMarkdownTree(cmd, "./docs/reference/")
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Printf("ERROR: %+v", err)
		os.Exit(1)
	}
}
