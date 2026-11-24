package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	melodyapplication "github.com/precision-soft/melody/v3/application"
	melodycontainer "github.com/precision-soft/melody/v3/container"
	containercontract "github.com/precision-soft/melody/v3/container/contract"
)

const (
	/* ServiceLedger is registered at the singleton lifetime: one instance for the whole process. */
	ServiceLedger = "demo.ledger"

	/* ServiceWorkStamp is registered at the scoped lifetime: one instance per scope, closed with it. */
	ServiceWorkStamp = "demo.work_stamp"
)

/* registerServices declares both lifetimes side by side, which is the point of the demo: the ledger outlives every unit of work, the stamp lives inside one of them. */
func registerServices(app *melodyapplication.Application) {
	app.RegisterService(
		ServiceLedger,
		func(resolver containercontract.Resolver) (*Ledger, error) {
			return NewLedger(), nil
		},
	)

	app.RegisterScopedService(
		ServiceWorkStamp,
		func(resolver containercontract.Resolver) (*WorkStamp, error) {
			/* the resolver handed to a scoped provider is the scope, so the process-wide ledger is reachable from inside it and stays a singleton */
			ledger := melodycontainer.MustFromResolverByType[*Ledger](resolver)

			return ledger.Open(), nil
		},
	)
}

/* Ledger is the process-wide record of what the scopes built and closed. It exists so the demo can prove a teardown instead of asking the reader to trust one, and it is registered as a singleton, so no scope closes it. */
type Ledger struct {
	mutex   sync.Mutex
	nextID  int
	created []string
	closed  []string
}

func NewLedger() *Ledger {
	return &Ledger{
		nextID:  0,
		created: make([]string, 0),
		closed:  make([]string, 0),
	}
}

/* processTag keeps the identifiers of two runs of the same binary apart, since each run builds its own ledger and would otherwise number its stamps from one again. */
func processTag() string {
	return strconv.Itoa(os.Getpid())
}

func (instance *Ledger) Open() *WorkStamp {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	instance.nextID = instance.nextID + 1

	stamp := &WorkStamp{
		id:     fmt.Sprintf("stamp-%s-%d", processTag(), instance.nextID),
		opened: time.Now(),
		ledger: instance,
	}

	instance.created = append(instance.created, stamp.id)

	return stamp
}

func (instance *Ledger) RecordClosed(id string) {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	instance.closed = append(instance.closed, id)
}

func (instance *Ledger) Created() []string {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	return append(make([]string, 0, len(instance.created)), instance.created...)
}

func (instance *Ledger) Closed() []string {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	return append(make([]string, 0, len(instance.closed)), instance.closed...)
}

/* WorkStamp is the scoped service: the container builds it on the first Get through a scope, hands the same instance to every later Get in that scope, and closes it when the scope closes. Close is the door the teardown looks for, so a value carrying Close() error is closed by the scope that built it. */
type WorkStamp struct {
	id     string
	opened time.Time
	ledger *Ledger
}

func (instance *WorkStamp) ID() string {
	return instance.id
}

func (instance *WorkStamp) OpenedAt() time.Time {
	return instance.opened
}

func (instance *WorkStamp) Close() error {
	instance.ledger.RecordClosed(instance.id)

	/* printed rather than collected: for a console run nothing of the application is left to render a report when the scope closes, so the record of the teardown is written as it happens */
	fmt.Println("[closed]", instance.id)

	return nil
}
