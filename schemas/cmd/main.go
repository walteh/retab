package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/walteh/retab/schemas"
)

func main() {
	ctx := context.Background()

	err := downloadAllJSONSchemas(ctx)
	if err != nil {
		panic(err)
	}
}

// Existing code...

func downloadAllJSONSchemas(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(schemas.KnownSchemas()))

	// Determine the current file's directory
	_, filename, _, _ := runtime.Caller(0)
	currentFileDir := filepath.Dir(filename)

	for _, schema := range schemas.KnownSchemas() {
		wg.Add(1)

		go func(schema string) {
			defer wg.Done()

			schemaData, err := schemas.DownloadJSONSchema(ctx, schema)
			if err != nil {
				errChan <- fmt.Errorf("Failed to download schema %s: %v", schema, err)
				return
			}

			// Generate relative path for the ../json/ directory
			destinationDir := filepath.Join(currentFileDir, "..", "json")
			destinationPath := filepath.Join(destinationDir, schema)

			// Create the directory if it doesn't exist
			err = os.MkdirAll(destinationDir, os.ModePerm)
			if err != nil {
				errChan <- fmt.Errorf("Failed to create directory: %v", err)
				return
			}

			// Write schema data to file
			err = os.WriteFile(destinationPath, schemaData, 0644)
			if err != nil {
				errChan <- fmt.Errorf("Failed to write schema to file: %v", err)
			}
		}(schema)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect any errors that occurred during downloading
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("Encountered multiple errors: %v", errs)
	}

	return nil
}
