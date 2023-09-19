package schemas

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
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
		return nil, fmt.Errorf("%s returned status code %d", schema, resp.StatusCode)
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return all, nil
}
