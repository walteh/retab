package file

import (
	"context"
	"os"

	"github.com/spf13/afero"
)

type aferoFileClient struct {
	fs afero.Fs
}

func NewAferoFile(fs afero.Fs) *aferoFileClient {
	return &aferoFileClient{
		fs: fs,
	}
}

func (mfs *aferoFileClient) Get(ctx context.Context, path string) (res []byte, err error) {
	// Read the content from the specified path
	res, err = afero.ReadFile(mfs.fs, path)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (mfs *aferoFileClient) Put(ctx context.Context, path string, data []byte) error {
	// Write the data to the specified path
	err := afero.WriteFile(mfs.fs, path, data, 0644)

	if err != nil {
		return err
	}

	return nil
}

func (af *aferoFileClient) AppendString(ctx context.Context, path string, data string) error {
	// Open the file in append mode, or create it if it doesn't exist
	file, err := af.fs.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the string data to the file
	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}

func (mfs *aferoFileClient) Delete(ctx context.Context, path string) error {
	// Delete the specified path
	err := mfs.fs.Remove(path)

	if err != nil {
		return err
	}

	return nil
}

func (mfs *aferoFileClient) GetFile(ctx context.Context, path string) (res afero.File, err error) {
	// Open the file at the specified path
	r, err := mfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	return r, nil
}
