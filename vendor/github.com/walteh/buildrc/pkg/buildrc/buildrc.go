package buildrc

import (
	"context"

	"github.com/walteh/buildrc/pkg/git"
)

type Buildrc struct {
	MajorRaw int `yaml:"major,flow" json:"major"`
}

func (me *Buildrc) Major() uint64 {
	return uint64(me.MajorRaw)
}

func LoadBuildrc(_ context.Context, _ git.GitProvider) (*Buildrc, error) {

	return &Buildrc{0}, nil
}
