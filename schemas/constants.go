package schemas

import (
	"strings"
)

const (
	registry = "json.schemastore.org"
)

var (
	regristryAliases = []string{
		"raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json",
	}
)

func SchemaRefName(str string) string {
	str = strings.TrimPrefix(str, "https://")
	str = strings.TrimPrefix(str, "http://")
	str = strings.TrimPrefix(str, registry)

	for _, alias := range regristryAliases {
		str = strings.TrimPrefix(str, alias)
	}

	str = strings.TrimPrefix(str, "/")

	return str
}

func KnownSchemas() []string {
	return knownSchemas
}

var knownSchemas = []string{
	"github-workflow.json",
}
