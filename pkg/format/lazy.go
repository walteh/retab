package format

import (
	"context"
	"io"
	"sync"
)

type LazyFormatProvider struct {
	provider     Provider
	providerOnce sync.Once
	providerFunc func() Provider
}

func (p *LazyFormatProvider) Format(ctx context.Context, cfg Configuration, reader io.Reader) (io.Reader, error) {
	p.providerOnce.Do(func() {
		p.provider = p.providerFunc()
	})

	return p.provider.Format(ctx, cfg, reader)
}

func NewLazyFormatProvider(providerFunc func() Provider) *LazyFormatProvider {
	return &LazyFormatProvider{
		providerFunc: providerFunc,
	}
}
