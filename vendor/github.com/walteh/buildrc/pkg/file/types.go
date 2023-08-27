package file

import (
	"context"

	"github.com/spf13/afero"
)

type FileAPI interface {
	Get(ctx context.Context, key string) (res []byte, err error)
	Put(ctx context.Context, key string, data []byte) error
	AppendString(ctx context.Context, key string, data string) error
	Delete(ctx context.Context, key string) error

	GetFile(ctx context.Context, key string) (res afero.File, err error)
}

type Client struct {
}
