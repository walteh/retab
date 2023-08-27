package bufwrite

import (
	"context"
	"io"

	"github.com/spf13/afero"
	"github.com/walteh/tftab/pkg/configuration"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
)

func Format(ctx context.Context, fls afero.Fs, path string, cfg configuration.Provider) (io.Reader, error) {
	fle, err := fls.Open(path)
	if err != nil {
		return nil, err
	}
	defer fle.Close()
	fileNode, err := parser.Parse(fle.Name(), fle, reporter.NewHandler(nil))
	if err != nil {
		return nil, err
	}

	read, write := io.Pipe()

	fmtr := newFormatter(write, fileNode, cfg)

	go func() {
		if err := fmtr.Run(); err != nil {
			err := write.CloseWithError(err)
			if err != nil {
				panic(err)
			}
			return
		} else {
			err := write.Close()
			if err != nil {
				panic(err)
			}
		}
	}()

	return read, nil
}

// // Format formats and writes the target module files into a read bucket.
// func Formatr(ctx context.Context, module bufmodule.Module) (_ storage.ReadBucket, retErr error) {

// 	// fileInfos, err := module.TargetFileInfos(ctx)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// readWriteBucket := storagemem.NewReadWriteBucket()
// 	// jobs := make([]func(context.Context) error, len(fileInfos))
// 	// for i, fileInfo := range fileInfos {
// 	// 	fileInfo := fileInfo
// 	// 	jobs[i] = func(ctx context.Context) (retErr error) {
// 	// 		moduleFile, err := module.GetModuleFile(ctx, fileInfo.Path())
// 	// 		if err != nil {
// 	// 			return err
// 	// 		}
// 	// 		defer func() {
// 	// 			retErr = multierr.Append(retErr, moduleFile.Close())
// 	// 		}()
// 	// 		fileNode, err := parser.Parse(moduleFile.ExternalPath(), moduleFile, reporter.NewHandler(nil))
// 	// 		if err != nil {
// 	// 			return err
// 	// 		}
// 	// 		writeObjectCloser, err := readWriteBucket.Put(ctx, moduleFile.Path())
// 	// 		if err != nil {
// 	// 			return err
// 	// 		}
// 	// 		defer func() {
// 	// 			retErr = multierr.Append(retErr, writeObjectCloser.Close())
// 	// 		}()

// 	// 		return writeObjectCloser.SetExternalPath(moduleFile.ExternalPath())
// 	// 	}
// 	// }
// 	// if err := thread.Parallelize(ctx, jobs); err != nil {
// 	// 	return nil, err
// 	// }
// 	return readWriteBucket, nil
// }

// func abc(ctx context.Context, module bufmodule.Module, fls afero.Fs, path string) (io.Reader, error) {

// 	// Note that external paths are set properly for the files in this read bucket.
// 	formattedReadBucket, err := bufformat.Format(ctx, module)
// 	if err != nil {
// 		return nil, err
// 	}

// 	reader := bytes.NewBuffer(nil)

// 	if err := storage.WalkReadObjects(
// 		ctx,
// 		formattedReadBucket,
// 		"",
// 		func(readObject storage.ReadObject) error {
// 			data, err := io.ReadAll(readObject)
// 			if err != nil {
// 				return err
// 			}
// 			if _, err := reader.Write(data); err != nil {
// 				return err
// 			}
// 			return nil
// 		},
// 	); err != nil {
// 		return nil, err
// 	}

// 	return reader, nil
// }
