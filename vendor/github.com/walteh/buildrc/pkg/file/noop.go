package file

import "context"

type NoopClient struct {
}

func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (m *NoopClient) Get(_ context.Context) (_ []byte, _ error) {
	return nil, nil
}

func (m *NoopClient) Put(_ context.Context, _ []byte) error {
	return nil
}
