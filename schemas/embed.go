package schemas

import "embed"

//go:embed json/*.json
var jsonSchemas embed.FS
