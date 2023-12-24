package resolvers

import (
	"os"

	"github.com/spf13/afero"
	"github.com/walteh/snake"
	"github.com/walteh/terrors"
)

func FSRunner() snake.Runner {
	return snake.GenRunResolver_In00_Out03(&FSResolver{})
}

type FSResolver struct {
	Dir  string `usage:"the directory to run in"`
	File string `usage:"the file to read the configuration from"`
}

func (me *FSResolver) Run() (afero.Fs, afero.File, error) {
	res := afero.NewOsFs()
	if me.Dir == "" {
		wrking, err := os.Getwd()
		if err != nil {
			return nil, nil, err
		}
		res = afero.NewBasePathFs(res, wrking)
	} else {
		res = afero.NewBasePathFs(res, me.Dir)

	}

	path := me.File

	if path == "" {
		path = "."
	}

	fle, err := res.Open(path)
	if err != nil {
		return res, nil, terrors.Wrap(err, "failed to open file")
	}

	return res, fle, nil
}
