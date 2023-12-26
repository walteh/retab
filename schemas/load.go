package schemas

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func LoadJSONSchema(ctx context.Context, str string) (*jsonschema.Schema, error) {

	ref := SchemaRefName(str)

	fmt.Println("ref", str, ref)

	if dat, err := jsonSchemas.ReadFile(filepath.Join("json", ref)); err == nil {
		// fmt.Println("schema2", string(dat))
		// return nil, err
		return jsonschema.CompileString(ref, string(dat))
	}

	dl, err := DownloadJSONSchema(ctx, str)
	if err != nil {
		return nil, err
	}

	return jsonschema.CompileString(ref, string(dl))
}
