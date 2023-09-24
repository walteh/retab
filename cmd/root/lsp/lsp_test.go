package lsp

import (
	"context"
	"testing"

	"github.com/walteh/retab/internal/debug"
)

func TestFull(t *testing.T) {

	ctx := debug.WithInstance(context.Background(), "./de.bug", "serve")

	t.Run("test", func(t *testing.T) {

		srv := NewServe()

		// write

		if err := srv.Run(ctx); err != nil {
			t.Fatal(err)
		}

	})
}
