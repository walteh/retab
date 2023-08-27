package proto

import (
	"bytes"
	"context"
	"io"

	"github.com/spf13/afero"

	"github.com/bufbuild/buf/private/buf/bufformat"
	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
	"github.com/bufbuild/buf/private/pkg/storage"
)

func ConvertProtoToHCL(ctx context.Context, fls afero.Fs, path string) (io.Reader, error) {
	module := NewModule(ctx, fls)
	return abc(ctx, module, fls, path)
}

func abc(ctx context.Context, module bufmodule.Module, fls afero.Fs, path string) (io.Reader, error) {

	// Note that external paths are set properly for the files in this read bucket.
	formattedReadBucket, err := bufformat.Format(ctx, module)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewBuffer(nil)

	if err := storage.WalkReadObjects(
		ctx,
		formattedReadBucket,
		"",
		func(readObject storage.ReadObject) error {
			data, err := io.ReadAll(readObject)
			if err != nil {
				return err
			}
			if _, err := reader.Write(data); err != nil {
				return err
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	return reader, nil
}
