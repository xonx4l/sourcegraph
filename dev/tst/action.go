package tst

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type actionManager struct {
	setup    []Action
	teardown []Action
}

type actionApplyCfg struct {
	test     *testing.T
	actions  []Action
	store    *scenarioStore
	failFast bool
}

type ActionResult interface {
	Get() any
}

type ActionFn func(ctx context.Context, store *scenarioStore) (ActionResult, error)

type Action interface {
	Name() string
	Hash() []byte
	Complete() bool
	Do(ctx context.Context, t *testing.T, store *scenarioStore) (ActionResult, error)
	String() string
}

type action struct {
	id       string
	name     string
	hash     []byte
	complete bool
	fn       ActionFn
}

func (a *action) Do(ctx context.Context, t *testing.T, store *scenarioStore) (ActionResult, error) {
	t.Helper()
	result, err := a.fn(ctx, store)
	a.complete = true
	return result, err
}

func (a *action) Hash() []byte {
	return a.hash
}

func (a *action) Name() string {
	return a.name
}

func (a *action) Complete() bool {
	return a.complete
}

func (a *action) String() string {
	return fmt.Sprintf("%s (%s)", a.name, a.id)
}

type actionResult[T any] struct {
	item T
}

func (a *actionResult[T]) Get() any {
	return a.item
}

func NewActionManager() *actionManager {
	return &actionManager{
		setup:    make([]Action, 0),
		teardown: make([]Action, 0),
	}
}

func (m *actionManager) AddSetup(actions ...Action) {
	m.setup = append(m.setup, actions...)
}
func (m *actionManager) AddTeardown(actions ...Action) {
	m.setup = append(m.teardown, actions...)
}

func (m *actionManager) setupPlan() string {
	b := strings.Builder{}
	for _, a := range m.setup {
		b.WriteString(a.String())
		b.WriteByte('\n')
	}

	return b.String()
}

func (m *actionManager) teardownPlan() string {
	b := strings.Builder{}
	actions := m.teardown
	for i := len(actions) - 1; i >= 0; i-- {
		b.WriteString(actions[i].String())
		b.WriteByte('\n')
	}

	return b.String()
}

func (m *actionManager) String() string {
	b := strings.Builder{}
	b.WriteString("Setup\n")
	b.WriteString("======\n")
	b.WriteString(m.setupPlan())
	b.WriteByte('\n')
	b.WriteString("Teardown\n")
	b.WriteString("========\n")
	b.WriteString(m.teardownPlan())
	return b.String()
}

func (m *actionManager) Apply(ctx context.Context, cfg *actionApplyCfg) error {
	var errs errors.MultiError
	for _, action := range cfg.actions {
		fmt.Printf("Applying '%s' = ", action)
		now := time.Now().UTC()

		var err error
		if !action.Complete() {
			_, err = action.Do(ctx, cfg.test, cfg.store)
		} else {
			fmt.Print("[SKIPPED]\n")
			continue
		}

		duration := time.Now().UTC().Sub(now)
		if err != nil {
			if cfg.failFast {
				fmt.Printf("[FAILED] (%s)\n", duration.String())
				return err
			} else {
				fmt.Printf("[FAILED] (%s)\n", duration.String())
				errs = errors.Append(errs, err)
			}
		} else {
			fmt.Printf("[SUCCESS] (%s)\n", duration.String())
		}
	}
	return errs
}
