package hclwrite

import (
	"context"
	"io"

	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
)

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewHclFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Targets() []string {
	return []string{"*.hcl", "*.tf", "*.tfvars", "*.hcl2"}
}

func (me *Formatter) Format(ctx context.Context, cfg configuration.Provider, read io.Reader) (io.Reader, error) {

	reads, err := io.ReadAll(read)
	if err != nil {
		return nil, err
	}

	err = checkErrors(ctx, reads, "")
	if err != nil {
		return nil, err
	}

	newContents, err := FormatBytes(cfg, reads)
	if err != nil {
		return nil, err
	}

	return newContents, nil
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
