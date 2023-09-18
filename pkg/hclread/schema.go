package hclread

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// load json or yaml schema file

func LoadJsonSchemaFile(ctx context.Context, document_uri string) (*jsonschema.Schema, error) {

	loader := jsonschema.NewCompiler()
	loader.Draft = jsonschema.Draft7

	resp, err := http.Get(document_uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status code %d", document_uri, resp.StatusCode)
	}

	// compile schema
	str, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	schema, err := jsonschema.CompileString(document_uri, string(str))
	if err != nil {
		return nil, err
	}

	return schema, nil
}
