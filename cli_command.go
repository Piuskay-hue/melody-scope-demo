package main

import (
	"fmt"

	melodyclicontract "github.com/precision-soft/melody/v3/cli/contract"
	melodycontainer "github.com/precision-soft/melody/v3/container"
	runtimecontract "github.com/precision-soft/melody/v3/runtime/contract"
)

/* ScopeReportCommand does what ScopeReportHandler does, through the console entry point. The command body creates no scope: the run made one and hands it over in the runtime. */
type ScopeReportCommand struct{}

func NewScopeReportCommand() *ScopeReportCommand {
	return &ScopeReportCommand{}
}

func (instance *ScopeReportCommand) Name() string {
	return "scope:report"
}

func (instance *ScopeReportCommand) Description() string {
	return "resolves the scoped service twice inside the run's scope and prints its identity"
}

func (instance *ScopeReportCommand) Flags() []melodyclicontract.Flag {
	return []melodyclicontract.Flag{}
}

func (instance *ScopeReportCommand) Run(
	runtimeInstance runtimecontract.Runtime,
	commandContext *melodyclicontract.CommandContext,
) error {
	first := melodycontainer.MustFromResolverByType[*WorkStamp](runtimeInstance.Scope())
	second := melodycontainer.MustFromResolverByType[*WorkStamp](runtimeInstance.Scope())

	ledger := melodycontainer.MustFromResolverByType[*Ledger](runtimeInstance.Container())

	writer := commandContext.Writer

	fmt.Fprintf(writer, "entry point:              %s\n", "cli")
	fmt.Fprintf(writer, "first resolution:         %s\n", first.ID())
	fmt.Fprintf(writer, "second resolution:        %s\n", second.ID())
	fmt.Fprintf(writer, "same instance in one run: %t\n", first == second)
	fmt.Fprintf(writer, "root container sees it:   %t\n", runtimeInstance.Container().Has(ServiceWorkStamp))
	fmt.Fprintf(writer, "created so far:           %v\n", ledger.Created())
	fmt.Fprintf(writer, "closed so far:            %v\n", ledger.Closed())

	return nil
}

var _ melodyclicontract.Command = (*ScopeReportCommand)(nil)
