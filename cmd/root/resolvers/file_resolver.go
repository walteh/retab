package resolvers

import (
	"github.com/spf13/afero"
	"github.com/walteh/snake"
)

func FileRunner() snake.Runner {
	return snake.GenRunResolver_In01_Out02(&FileResolver{})
}

type FileResolver struct {
	File string `usage:"The file to read the configuration from."`
}

func (me *FileResolver) Run(fls afero.Fs) (afero.File, error) {
	return fls.Open(me.File)
}
