package file

import "github.com/spf13/afero"

var _ FileAPI = (*OSFile)(nil)

type OSFile struct {
	*aferoFileClient
}

func NewOSFile() *OSFile {
	return &OSFile{
		aferoFileClient: NewAferoFile(afero.NewOsFs()),
	}
}
