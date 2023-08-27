package file

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func Sha256(ctx context.Context, fls afero.Fs, path string) (afero.File, error) {

	// Get the filename
	name := filepath.Base(path)
	hname := fmt.Sprintf("%s.sha256", name)

	fle, err := fls.Create(path + ".sha256")
	if err != nil {
		return nil, fmt.Errorf("error creating SHA-256 checksum file %s: %v", hname, err)
	}

	// Open the file
	file, err := fls.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Compute the SHA-256 checksum
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("error computing SHA-256 checksum: %v", err)
	}
	hashOutput := hasher.Sum(nil)

	zerolog.Ctx(ctx).Debug().Msgf("computed SHA-256 checksum for %s", hname)

	// Write the checksum and the filename to a file
	_, err = fle.WriteString(fmt.Sprintf("%x  %s", hashOutput, filepath.Base(name)))
	if err != nil {
		return nil, fmt.Errorf("error writing SHA-256 checksum file %s: %v", hname, err)
	}

	return fle, nil

}
