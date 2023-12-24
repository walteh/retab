# snake 🐍

> [!CAUTION]
> This is a personal project I use to simplify my own tool development. It is not intended for public use. It is not well tested. It is not well documented. It is not well maintained. It is not well anything. Use at your own risk.<<

## Description

Snake is a Go library designed to help build tools faster. It combines the high-level binding logic of the Kong CLI library with the Cobra CLI library, providing a powerful and efficient way to create custom golang tools.

It is designed to be built on top of existing frameworks. Right now, it only supports Cobra.

From a Cobra perspective, think of it like this: Cobra is a library that helps you build a CLI. Snake is a library that helps you manage interdependencies between commands to make it easier to build a CLI.

## Installation

To install the Snake library, use the `go get` command:

```bash
go get github.com/walteh/snake
```

## Building Resolvers

Define a new resolver:

```go
package custom

func Runner() snake.Runner {
	return snake.GenRunResolver_In00_Out02(&CustomResolver{})
}

type CustomInterface interface {
}

type CustomResolver struct {
}

func (me *CustomResolver) Run() (CustomInterface, error) {
	return struct{}{}, nil
}
```

## Building Commands
```go
package basic 

import (
    "fmt"
    "xxx/custom"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In01_Out02(&Handler{})
}

type Handler struct {
	Value string `default:"default"`
}

func (me *Handler) Name() string {
	return "basic"
}

func (me *Handler) Description() string {
	return "basic description"
}

func (me *Handler) Run(dat custom.CustomInterface) (snake.Output, error) {
	return &snake.RawTextOutput{
		Data: fmt.Sprintf("hello %s, my value is %s", me.Value, dat),
	}, nil
}
```

## Building a CLI

```go
package main

import (
    "context"
    "github.com/yourusername/snake"
    "github.com/yourusername/snake/scobra"
    "github.com/yourusername/snake/examples/basic"
    "github.com/yourusername/snake/examples/custom"
    "github.com/spf13/cobra"
)

func main() {
    ctx := context.Background()

    cmd, scmd, err := NewCommand(ctx)
    if err != nil {
        panic(err)
    }

    // run as a normal cobra command
    if err := cmd.Execute(); err != nil {
        panic(err)
    }

    // or run with some common cli error handling 
    scobra.ExecuteHandlingError(ctx, scmd)

}

func NewCommand(ctx context.Context) (*cobra.Command, *scobra.CobraSnake, error) {

	cmd := &cobra.Command{
		Use: "root",
	}

	impl := scobra.NewCobraSnake(ctx, cmd)

	opts := snake.Opts(
		snake.Commands(
			snake.Command(basic.Runner, impl, &cobra.Command{}),
		),
		snake.Resolvers(
			snake.Resolver(custom.Runner),
		),
	)

	_, err := snake.NewSnakeWithOpts(ctx, impl, opts)
	if err != nil {
		return err
	}

	return cmd, impl, err
}
```