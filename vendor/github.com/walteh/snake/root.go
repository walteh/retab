package snake

import (
	"context"
	"sync"
)

const RootCommandName = "______root_____________"

type rootKeyT struct {
}

var rootKey = rootKeyT{}

func SetRootCommand(ctx context.Context, cmd *NamedCommand) context.Context {
	return context.WithValue(ctx, rootKey, cmd)
}

func GetRootCommand(ctx context.Context) *NamedCommand {
	p, ok := ctx.Value(rootKey).(*NamedCommand)
	if ok {
		return p
	}
	return nil
}

type namedCommandKeyT struct {
}

var namedCommandKey = namedCommandKeyT{}

type namedCommandMap map[string]*NamedCommand

var namedCommandMutex = sync.RWMutex{}

func SetNamedCommand(ctx context.Context, name string, cmd *NamedCommand) context.Context {

	ncm, ok := ctx.Value(namedCommandKey).(namedCommandMap)
	if !ok {
		ncm = make(namedCommandMap)
	}
	namedCommandMutex.Lock()
	ncm[name] = cmd
	namedCommandMutex.Unlock()

	return context.WithValue(ctx, namedCommandKey, ncm)
}

func GetNamedCommand(ctx context.Context, name string) *NamedCommand {
	p, ok := ctx.Value(namedCommandKey).(namedCommandMap)
	if ok {
		namedCommandMutex.RLock()
		defer namedCommandMutex.RUnlock()
		return p[name]
	}
	return nil
}

func GetAllNamedCommands(ctx context.Context) namedCommandMap {
	p, ok := ctx.Value(namedCommandKey).(namedCommandMap)
	if ok {
		namedCommandMutex.RLock()
		defer namedCommandMutex.RUnlock()
		return p
	}
	return nil
}

type activeCommandKeyT struct {
}

var activeCommandKey = activeCommandKeyT{}

func SetActiveCommand(ctx context.Context, str string) context.Context {
	return context.WithValue(ctx, activeCommandKey, str)
}

func GetActiveCommand(ctx context.Context) string {
	p, ok := ctx.Value(activeCommandKey).(string)
	if ok {
		return p
	}
	return ""
}

func ClearActiveCommand(ctx context.Context) context.Context {
	return context.WithValue(ctx, activeCommandKey, "")
}

func GetActiveNamedCommand(ctx context.Context) *NamedCommand {
	return GetNamedCommand(ctx, GetActiveCommand(ctx))
}
