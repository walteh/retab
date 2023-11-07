package file

import (
	"bytes"
	"context"
	errz "errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"slices"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-faster/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func Diff(ctx context.Context, fls afero.Fs, baseDir string, compareDir string, globs []string) ([]string, error) {

	baseFs := afero.NewBasePathFs(fls, baseDir)
	compareFs := afero.NewBasePathFs(fls, compareDir)

	if baseDir == "" || baseDir == "." {
		baseFs = fls
	}
	if compareDir == "" || compareDir == "." {
		compareFs = fls
	}

	baseIoFs := afero.NewIOFS(baseFs)
	compareIoFs := afero.NewIOFS(compareFs)

	baseMap := make(map[string]bool)
	compareMap := make(map[string]bool)

	for _, glob := range globs {

		// Read directory contents
		tfiles1, err := doublestar.Glob(baseIoFs, glob)
		if err != nil {
			return nil, errors.Errorf("Error reading directory: %v", err)
		}

		for _, file := range tfiles1 {
			baseMap[filepath.Clean(file)] = true
		}

		tfiles2, err := doublestar.Glob(compareIoFs, glob)
		if err != nil {
			return nil, errors.Errorf("Error reading directory: %v", err)
		}

		for _, file := range tfiles2 {
			compareMap[filepath.Clean(file)] = true
		}
	}

	// we could check the length here, but we take the extra step of sorting the arrays for more readable output

	baseFileArr := []string{}
	compareFileARR := []string{}

	for file := range baseMap {
		baseFileArr = append(baseFileArr, file)
	}

	for file := range compareMap {
		compareFileARR = append(compareFileARR, file)
	}

	slices.Sort(baseFileArr)
	slices.Sort(compareFileARR)

	// if the arrays are not the same length or the contents are not the same, we can return the diff
	if len(baseFileArr) != len(compareFileARR) || !slices.Equal(baseFileArr, compareFileARR) {
		return sliceDiff(ctx, baseFileArr, compareFileARR), nil
	}

	zerolog.Ctx(ctx).Debug().Msg("computing concurrent folder diff")

	// at this point we now the file arrays are the same, so we compare the filesystems concurrently
	diff, err := concurrentFolderDiff(ctx, baseFs, compareFs, baseFileArr)
	if err != nil {
		return nil, err
	}

	zerolog.Ctx(ctx).Debug().Any("diff", diff).Msg("done computing concurrent folder diff")

	return diff, nil
}

// readAndCompareFiles reads the content of the two files and checks if they are identical
func readAndCompareFiles(ctx context.Context, baseFs afero.Fs, compareFs afero.Fs, file string) (bool, error) {

	// Open files
	baseContent, err := baseFs.Open(file)
	if err != nil {
		log.Printf("Error reading files: %v", err)
		return false, err
	}
	defer baseContent.Close()

	compareContent, err := compareFs.Open(file)
	if err != nil {
		log.Printf("Error reading files: %v", err)
		return false, err
	}

	defer compareContent.Close()

	// make sure they are not directories
	baseFileStat, err := baseContent.Stat()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("file", file).Str("group", "base").Msg("problem getting file info")
		return false, err
	}

	compareFileStat, err := compareContent.Stat()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("file", file).Str("group", "compare").Msg("problem getting file info")
		return false, err
	}

	if baseFileStat.IsDir() || compareFileStat.IsDir() {
		return baseFileStat.IsDir() == compareFileStat.IsDir(), nil
	}

	// Compare file sizes
	if baseFileStat.Size() != compareFileStat.Size() {
		return false, err
	}

	// Compare file contents
	baseFileBytes, err := io.ReadAll(baseContent)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("file", file).Str("group", "base").Msg("problem reading file")
		return false, err
	}

	compareFileBytes, err := io.ReadAll(compareContent)
	if err != nil {
		log.Printf("Error reading files: %v", err)
		return false, err
	}

	return bytes.Equal(baseFileBytes, compareFileBytes), nil
}

func sliceDiff(ctx context.Context, baseArr, compareArr []string) []string {

	zerolog.Ctx(ctx).Info().Int("base", len(baseArr)).Int("compare", len(compareArr)).Msg("computing slice diff")

	diffs := []string{}

	// Using a map for faster lookup
	fileMap := make(map[string]bool)

	// Populate map with files from files1
	for _, file := range baseArr {
		fileMap[file] = true
	}

	// Check for missing or extra files in files2
	for _, file := range compareArr {
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

	grp := sync.WaitGroup{}

	diffs := []string{}

	mutex := sync.Mutex{}

	errs := []error{}

	// Send files to channels
	for _, file := range files {
		grp.Add(1)
		go func(file string) {
			defer grp.Done()
			ok, err := readAndCompareFiles(ctx, flsa, flsb, file)
			mutex.Lock()
			defer mutex.Unlock()
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Str("file", file).Msg("problem reading file")
				errs = append(errs, err)
				return
			}
			if !ok {
				diffs = append(diffs, fmt.Sprintf("~ %s", file))
			}

		}(file)
	}

	grp.Wait()

	return diffs, errz.Join(errs...)
}
