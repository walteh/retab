package main

import (
	"context"
	"log"

	"github.com/walteh/retab/cmd/root"
)

func main() {

	ctx := context.Background()

	cmd, err := root.NewCommand()
	if err != nil {
		log.Fatalf("ERROR: %+v", err)
	}

	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatalf("ERROR: %+v", err)
	}
}
