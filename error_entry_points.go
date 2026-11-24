package main

import (
	"errors"
	"fmt"
	nethttp "net/http"

	melodyclicontract "github.com/precision-soft/melody/v3/cli/contract"
	melodycontainer "github.com/precision-soft/melody/v3/container"
	httpcontract "github.com/precision-soft/melody/v3/http/contract"
	runtimecontract "github.com/precision-soft/melody/v3/runtime/contract"
)

const scopeFailFlagBrokenClose = "broken-close"

/* ScopeFailHandler resolves the scoped service and then fails, to show what the request's scope does with a service it already built when the handler answers an error instead of a response. */
func ScopeFailHandler(
	runtimeInstance runtimecontract.Runtime,
	writer nethttp.ResponseWriter,
	request httpcontract.Request,
) (httpcontract.Response, error) {
	stamp := melodycontainer.MustFromResolverByType[*WorkStamp](runtimeInstance.Scope())

	return nil, fmt.Errorf("demo: handler failed after resolving %s", stamp.ID())
}

/* ScopeFailCommand does the same from the console. With --broken-close it also resolves a service that refuses to close, so the command's own error and the scope's teardown error have to be reported together. */
type ScopeFailCommand struct{}

func NewScopeFailCommand() *ScopeFailCommand {
	return &ScopeFailCommand{}
}

func (instance *ScopeFailCommand) Name() string {
	return "scope:fail"
}

func (instance *ScopeFailCommand) Description() string {
	return "fails on purpose after resolving a scoped service, to show the teardown on the error path"
}

func (instance *ScopeFailCommand) Flags() []melodyclicontract.Flag {
	return []melodyclicontract.Flag{
		&melodyclicontract.BoolFlag{
			Name:  scopeFailFlagBrokenClose,
			Usage: "also resolve a scoped service whose Close refuses",
		},
	}
}

func (instance *ScopeFailCommand) Run(
	runtimeInstance runtimecontract.Runtime,
	commandContext *melodyclicontract.CommandContext,
) error {
	stamp := melodycontainer.MustFromResolverByType[*WorkStamp](runtimeInstance.Scope())

	fmt.Fprintf(commandContext.Writer, "resolved: %s\n", stamp.ID())

	if true == commandContext.Bool(scopeFailFlagBrokenClose) {
		melodycontainer.MustFromResolverByType[*Fragile](runtimeInstance.Scope())

		fmt.Fprintln(commandContext.Writer, "resolved: a service that refuses to close")
	}

	return errors.New("demo: command failed after resolving the scoped services")
}

var _ melodyclicontract.Command = (*ScopeFailCommand)(nil)
