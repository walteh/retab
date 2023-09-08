package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"slices"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/afero"
)

func Diff(ctx context.Context, fls afero.Fs, f1 string, f2 string, globs []string) ([]string, error) {
	fls1 := afero.NewBasePathFs(fls, f1)
	fls2 := afero.NewBasePathFs(fls, f2)

	files1 := []string{}
	files2 := []string{}

	for _, glob := range globs {

		// Read directory contents
		tfiles1, err := doublestar.Glob(afero.NewIOFS(fls1), glob, doublestar.WithFilesOnly())
		if err != nil {
			log.Fatalf("Error reading directory: %v", err)
		}

		files1 = append(files1, tfiles1...)

		tfiles2, err := doublestar.Glob(afero.NewIOFS(fls2), glob, doublestar.WithFilesOnly())
		if err != nil {
			log.Fatalf("Error reading directory: %v", err)
		}

		files2 = append(files2, tfiles2...)
	}

	// sort and compare
	slices.Sort(files1)
	slices.Sort(files2)

	// Compare lists of files
	if !slices.Equal(files1, files2) {
		return sliceDiff(files1, files2), nil
	}

	return concurrentFolderDiff(ctx, fls1, fls2, files1)
}

const chunkSize = 64000

// readAndCompareFiles reads the content of the two files and checks if they are identical
func readAndCompareFiles(flsa afero.Fs, flsb afero.Fs, file string) bool {

	// Open files
	content1, erra := flsa.Open(file)
	if erra != nil {
		log.Printf("Error reading files: %v", erra)
		return false
	}
	defer content1.Close()

	content2, errb := flsb.Open(file)
	if errb != nil {
		log.Printf("Error reading files: %v", errb)
		return false
	}

	defer content2.Close()

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := content1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := content2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				return false
			} else {
				log.Fatal(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}

func sliceDiff(a, b []string) []string {

	diffs := []string{}

	// Using a map for faster lookup
	fileMap := make(map[string]bool)

	// Populate map with files from files1
	for _, file := range a {
		fileMap[file] = true
	}

	// Check for missing or extra files in files2
	for _, file := range b {
		if _, found := fileMap[file]; found {
			// remove from map if found in files2
			delete(fileMap, file)
		} else {
			// extra file in files2
			diffs = append(diffs, fmt.Sprintf("- %s", file))
		}
	}

	// Remaining files in map are missing in files2
	for file := range fileMap {
		diffs = append(diffs, fmt.Sprintf("+ %s", file))
	}

	return diffs
}

func concurrentFolderDiff(ctx context.Context, flsa afero.Fs, flsb afero.Fs, files []string) ([]string, error) {

	// Create channels for files and diffs
	diffsChan := make(chan []string)

	grp := sync.WaitGroup{}

	diffs := []string{}

	go func() {
		// Collect diffs
		for d := range diffsChan {
			diffs = append(diffs, d...)
		}
	}()

	// Send files to channels
	for _, file := range files {
		grp.Add(1)
		go func(file string) {
			defer grp.Done()
			if !readAndCompareFiles(flsa, flsb, file) {
				diffsChan <- []string{fmt.Sprintf("~ %s", file)}
			}

		}(file)
	}

	grp.Wait()

	// Close channel
	close(diffsChan)

	// Return diffs
	return diffs, nil
}
