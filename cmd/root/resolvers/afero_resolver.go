package resolvers

import (
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/walteh/snake"
)

var _ snake.Flagged = (*AferoResolver)(nil)

type AferoResolver struct {
}

func (me *AferoResolver) Flags(flgs *pflag.FlagSet) {
}

func (me *AferoResolver) Run() (afero.Fs, error) {
	return afero.NewOsFs(), nil
}
