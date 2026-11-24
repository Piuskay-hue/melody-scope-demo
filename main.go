package main

import (
	melodyapplication "github.com/precision-soft/melody/v3/application"
	melodylogging "github.com/precision-soft/melody/v3/logging"
)

/* main boots one Melody application and lets its own arguments choose the entry point: no command means the http server, any command name means the console. Both entries reach the registrations below. */
func main() {
	ctx, stop := melodyapplication.NewSignalContext()
	defer stop()

	/* NewApplication and Run recover their own panics; a refusal raised between them, by the registration below, reaches no such recovery and would leave a bare stack trace. This records it and exits 1. */
	defer func() {
		melodylogging.LogOnRecoverAndExit(melodylogging.EmergencyLogger(), recover(), 1)
	}()

	app := melodyapplication.NewApplication(ctx, nil, nil)

	registerServices(app)
	registerErrorPathServices(app)

	app.RegisterHttpRoute("GET", "/scope-report", ScopeReportHandler)
	app.RegisterHttpRoute("GET", "/scope-fail", ScopeFailHandler)

	app.RegisterCliCommand(NewScopeReportCommand())
	app.RegisterCliCommand(NewScopeFailCommand())

	app.Run()
}
