package resolvers

import (
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/walteh/snake"
)

func FSRunner() snake.Runner {
	return snake.GenRunResolver_In00_Out02(&FSResolver{})
}

type FSResolver struct {
}

func (me *FSResolver) Flags(_ *pflag.FlagSet) {

}

func (me *FSResolver) Run() (afero.Fs, error) {
	return afero.NewOsFs(), nil
}
