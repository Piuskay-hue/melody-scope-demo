package main

import (
	"errors"

	melodyapplication "github.com/precision-soft/melody/v3/application"
	containercontract "github.com/precision-soft/melody/v3/container/contract"
)

const (
	/* ServiceFragile is a scoped service whose Close refuses, so the teardown error path is observable. */
	ServiceFragile = "demo.fragile"
)

/* registerErrorPathServices adds the one service the failing entry points resolve. It is registered only for that purpose: a service that refuses to close is a wiring bug everywhere else. */
func registerErrorPathServices(app *melodyapplication.Application) {
	app.RegisterScopedService(
		ServiceFragile,
		func(resolver containercontract.Resolver) (*Fragile, error) {
			return &Fragile{}, nil
		},
	)
}

/* Fragile answers an error from Close, which the scope reports as its own teardown error. */
type Fragile struct{}

func (instance *Fragile) Close() error {
	return errors.New("demo: this service refused to close")
}
