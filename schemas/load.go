package schemas

import (
	"context"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/walteh/terrors"
)

func LoadJSONSchema(ctx context.Context, str string) (*jsonschema.Schema, error) {

	ref := SchemaRefName(str)

	if dat, err := jsonSchemas.ReadFile(filepath.Join("json", ref)); err == nil {
		ljson, err := jsonschema.CompileString(ref, string(dat))
		if err != nil {
			return nil, terrors.Wrap(err, "invalid schema - failed to compile")
		}
		return ljson, nil
	}

	dl, err := DownloadJSONSchema(ctx, str)
	if err != nil {
		return nil, err
	}

	return jsonschema.CompileString(ref, string(dl))
}
