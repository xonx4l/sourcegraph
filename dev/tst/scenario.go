package tst

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/dev/tst/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Scenario interface {
	append(a ...*actionV2)
	Plan() string
	Apply(ctx context.Context) error
	Teardown(ctx context.Context) error
}

type GithubScenarioV2 struct {
	id               string
	t                *testing.T
	client           *GitHubClient
	actions          []*actionV2
	reporter         Reporter
	appliedActionIdx int
}

var _ Scenario = (*GithubScenarioV2)(nil)

func NewGithubScenarioV2(ctx context.Context, t *testing.T, cfg config.Config) (*GithubScenarioV2, error) {
	client, err := NewGitHubClient(ctx, cfg.GitHub)
	if err != nil {
		return nil, err
	}
	uid := []byte(uuid.NewString())
	id := base64.RawStdEncoding.EncodeToString(uid[:])[:10]
	return &GithubScenarioV2{
		id:       id,
		t:        t,
		client:   client,
		actions:  make([]*actionV2, 0),
		reporter: NoopReporter{},
	}, nil
}

func (s *GithubScenarioV2) Verbose() {
	s.reporter = &ConsoleReporter{}
}

func (s *GithubScenarioV2) Quiet() {
	s.reporter = NoopReporter{}
}

func (s *GithubScenarioV2) append(actions ...*actionV2) {
	s.actions = append(s.actions, actions...)
}

func (s *GithubScenarioV2) Plan() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Scenario %q\n", s.id)
	sb.WriteString("== Setup ==\n")
	for _, action := range s.actions {
		fmt.Fprintf(sb, "- %s\n", action.name)
	}
	sb.WriteString("== Teardown ==\n")
	for _, action := range reverse(s.actions) {
		if action.teardown == nil {
			continue
		}
		fmt.Fprintf(sb, "- %s\n", action.name)
	}
	return sb.String()
}

func (s *GithubScenarioV2) IsApplied() bool {
	return s.appliedActionIdx >= len(s.actions)
}

func (s *GithubScenarioV2) Apply(ctx context.Context) error {
	s.t.Helper()
	s.t.Cleanup(func() { s.Teardown(ctx) })
	var errs errors.MultiError
	setup := s.actions
	failFast := true

	if s.appliedActionIdx >= len(s.actions) {
		return errors.New("all actions already applied")
	}

	start := time.Now()
	for i, action := range setup {
		now := time.Now().UTC()

		var err error
		if s.appliedActionIdx <= i {
			s.reporter.Writef("(Setup) Applying '%s' = ", action.name)
			err = action.apply(ctx)
			s.appliedActionIdx++
		} else {
			s.reporter.Writef("(Setup) Skipping '%s' \n", action.name)
			continue
		}

		duration := time.Now().UTC().Sub(now)
		if err != nil {
			if failFast {
				s.reporter.Writef("[FAILED] (%s)\n", duration.String())
				return err
			} else {
				s.reporter.Writef("[FAILED] (%s)\n", duration.String())
				errs = errors.Append(errs, err)
			}
		} else {
			s.reporter.Writef("[SUCCESS] (%s)\n", duration.String())
		}
	}

	s.reporter.Writef("Setup complete in %s\n\n", time.Now().UTC().Sub(start))
	return errs
}

func (s *GithubScenarioV2) Teardown(ctx context.Context) error {
	s.t.Helper()
	var errs errors.MultiError
	teardown := reverse(s.actions)
	failFast := false

	start := time.Now()
	for _, action := range teardown {
		if action.teardown == nil {
			continue
		}
		now := time.Now().UTC()

		s.reporter.Writef("(Teardown) Applying '%s' = ", action.name)
		err := action.teardown(ctx)
		duration := time.Now().UTC().Sub(now)

		if err != nil {
			if failFast {
				s.reporter.Writef("[FAILED] (%s)\n", duration.String())
				return err
			}
			s.reporter.Writef("[FAILED] (%s)\n", duration.String())
			errs = errors.Append(errs, err)
		} else {
			s.reporter.Writef("[SUCCESS] (%s)\n", duration.String())
		}
	}
	s.reporter.Writef("Teardown complete in %s\n", time.Now().UTC().Sub(start))
	return errs
}

func (s *GithubScenarioV2) CreateOrg(name string) *Org {
	baseOrg := &Org{
		s:    s,
		name: name,
	}

	createOrg := &actionV2{
		name: "org:create:" + name,
		apply: func(ctx context.Context) error {
			orgName := fmt.Sprintf("org-%s-%s", name, s.id)
			org, err := s.client.createOrg(ctx, orgName)
			if err != nil {
				return err
			}
			baseOrg.name = org.GetLogin()
			return nil
		},
		teardown: func(context.Context) error {
			deleteURL := fmt.Sprintf("https://ghe.sgdev.org/organizations/%s/settings/profile", baseOrg.name)
			fmt.Printf("Visit %q to delete the org\n", deleteURL)
			return nil
		},
	}

	s.append(createOrg)
	return baseOrg
}
