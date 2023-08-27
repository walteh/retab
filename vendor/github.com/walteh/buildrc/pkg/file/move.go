package file

import (
	"context"

	"github.com/spf13/afero"
)

func Move(ctx context.Context, fls afero.Fs, src, dst string) error {
	if err := fls.Rename(src, dst); err != nil {
		return err
	}
	return nil
}
