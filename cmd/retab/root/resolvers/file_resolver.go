package resolvers

import (
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/walteh/snake"
)

var _ snake.Flagged = (*FileResolver)(nil)

type FileResolver struct {
	File string
}

func (me *FileResolver) Flags(p *pflag.FlagSet) {
	p.StringVarP(&me.File, "file", "f", "", "path of the file/directory to process")
}

func (me *FileResolver) Run(fls afero.Fs) (afero.File, error) {

	return fls.Open(me.File)
}
