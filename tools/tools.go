package tools

//go:generate go build  -mod=vendor -o ./bin/task github.com/go-task/task/v3/cmd/task
//go:generate ./bin/task tools

import (
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/go-task/task/v3/cmd/task"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/vektra/mockery/v2"
	_ "gotest.tools/gotestsum"
)
