package document

import "github.com/spf13/afero"

func (me *Document) AsFile() (afero.File, error) {
	fle, err := afero.NewMemMapFs().Open(me.Filename)
	if err != nil {
		return nil, err
	}

	_, err = fle.Write(me.Text)
	if err != nil {
		return nil, err
	}

	return fle, nil
}
