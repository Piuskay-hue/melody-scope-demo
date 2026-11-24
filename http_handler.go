package main

import (
	nethttp "net/http"

	melodycontainer "github.com/precision-soft/melody/v3/container"
	melodyhttp "github.com/precision-soft/melody/v3/http"
	httpcontract "github.com/precision-soft/melody/v3/http/contract"
	runtimecontract "github.com/precision-soft/melody/v3/runtime/contract"
)

/* ScopeReport is the body both entry points answer, so their outputs can be read side by side. */
type ScopeReport struct {
	EntryPoint string   `json:"entryPoint"`
	Route      string   `json:"route,omitempty"`
	First      string   `json:"firstResolution"`
	Second     string   `json:"secondResolution"`
	SameInWork bool     `json:"sameInstanceWithinOneUnitOfWork"`
	RootSees   bool     `json:"rootContainerSeesScopedService"`
	Created    []string `json:"createdSoFar"`
	Closed     []string `json:"closedSoFar"`
}

/* ScopeReportHandler resolves the scoped service twice in the request's own scope. The kernel made that scope before the handler ran; the runtime carries it here. */
func ScopeReportHandler(
	runtimeInstance runtimecontract.Runtime,
	writer nethttp.ResponseWriter,
	request httpcontract.Request,
) (httpcontract.Response, error) {
	first := melodycontainer.MustFromResolverByType[*WorkStamp](runtimeInstance.Scope())
	second := melodycontainer.MustFromResolverByType[*WorkStamp](runtimeInstance.Scope())

	/* the ledger is a singleton: it is reached through the container, not through the request's scope */
	ledger := melodycontainer.MustFromResolverByType[*Ledger](runtimeInstance.Container())

	return melodyhttp.JsonResponse(
		nethttp.StatusOK,
		ScopeReport{
			EntryPoint: "http",
			Route:      request.RoutePattern(),
			First:      first.ID(),
			Second:     second.ID(),
			SameInWork: first == second,
			RootSees:   runtimeInstance.Container().Has(ServiceWorkStamp),
			Created:    ledger.Created(),
			Closed:     ledger.Closed(),
		},
	)
}
