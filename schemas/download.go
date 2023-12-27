package schemas

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/walteh/terrors"
)

func DownloadJSONSchema(ctx context.Context, schema string) ([]byte, error) {
	if filepath.Base(schema) == schema {
		schema = fmt.Sprintf("https://%s/%s", registry, schema)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, schema, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, terrors.Errorf("%s returned status code %d", schema, resp.StatusCode)
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return all, nil
}

func DownloadAllJSONSchemas(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(KnownSchemas()))

	// Determine the current file's directory
	_, filename, _, _ := runtime.Caller(0)
	currentFileDir := filepath.Dir(filename)

	for _, schema := range KnownSchemas() {
		wg.Add(1)

		go func(schema string) {
			defer wg.Done()

			schemaData, err := DownloadJSONSchema(ctx, schema)
			if err != nil {
				errChan <- terrors.Errorf("Failed to download schema %s: %v", schema, err)
				return
			}

			// Generate relative path for the ../json/ directory
			destinationDir := filepath.Join(currentFileDir, "..", "json")
			destinationPath := filepath.Join(destinationDir, schema)

			// Create the directory if it doesn't exist
			err = os.MkdirAll(destinationDir, os.ModePerm)
			if err != nil {
				errChan <- terrors.Errorf("Failed to create directory: %v", err)
				return
			}

			// Write schema data to file
			err = os.WriteFile(destinationPath, schemaData, 0644)
			if err != nil {
				errChan <- terrors.Errorf("Failed to write schema to file: %v", err)
			}
		}(schema)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect any errors that occurred during downloading
	errs := make([]error, 0)
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return terrors.Errorf("Encountered multiple errors: %v", errs)
	}

	return nil
}
