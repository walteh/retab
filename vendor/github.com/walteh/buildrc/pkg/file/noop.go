package file

import "context"

type NoopClient struct {
}

func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (m *NoopClient) Get(ctx context.Context) (res []byte, err error) {
	return nil, nil
}

func (m *NoopClient) Put(ctx context.Context, data []byte) error {
	return nil
}
