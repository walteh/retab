package schemas

import (
	"context"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func LoadJSONSchema(ctx context.Context, str string) (*jsonschema.Schema, error) {

	ref := SchemaRefName(str)

	if dat, err := jsonSchemas.ReadFile(filepath.Join("json", ref)); err == nil {
		return jsonschema.CompileString(ref, string(dat))
	}

	dl, err := DownloadJSONSchema(ctx, str)
	if err != nil {
		return nil, err
	}

	return jsonschema.CompileString(ref, string(dl))
}
