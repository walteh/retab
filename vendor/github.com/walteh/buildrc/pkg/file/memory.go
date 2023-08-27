package file

import (
	"github.com/spf13/afero"
)

var _ FileAPI = (*MemoryFileClient)(nil)

type MemoryFileClient struct {
	*aferoFileClient
}

func NewMemoryFile() *MemoryFileClient {
	return &MemoryFileClient{
		aferoFileClient: NewAferoFile(afero.NewMemMapFs()),
	}
}
