package snake

import (
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	root = Ctx{
		bindings:  make(map[string]*reflect.Value),
		resolvers: make(map[string]Method),
		cmds:      make(map[string]*cobra.Command),
	}
)

type Flagged interface {
	Flags(*pflag.FlagSet)
}

func NewArgument[I any](method Flagged) {
	_ = NewArgContext[I](&root, method)
}

func NewCmd(cmd *cobra.Command, method Flagged) {
	_ = NewCmdContext(&root, cmd.Name(), cmd, method)
}

func NewCmdContext(con *Ctx, name string, cbra *cobra.Command, m Flagged) Method {

	ec := &method{
		flags:              m.Flags,
		validationStrategy: commandResponseValidationStrategy,
		responseStrategy:   commandResponseHandleStrategy,
		name:               prefix_command + name,
		method:             getRunMethod(m),
	}

	con.runlock.Lock()
	defer con.runlock.Unlock()

	con.cmds[name] = cbra

	con.resolvers[ec.name] = ec

	return ec
}

func methodName(typ reflect.Type) string {
	return prefix_argument + typ.String()
}

func NewArgContext[I any](con *Ctx, m Flagged) Method {

	ec := &method{
		flags:              m.Flags,
		validationStrategy: validateArgumentResponse[I],
		responseStrategy:   handleArgumentResponse[I],
		name:               methodName(reflect.TypeOf((*I)(nil)).Elem()),
		method:             getRunMethod(m),
	}

	con.runlock.Lock()
	defer con.runlock.Unlock()

	con.resolvers[ec.name] = ec

	return ec
}
