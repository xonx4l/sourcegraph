package tst

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/dev/tst/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Scenario interface {
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

func NewGithubScenarioV2(t *testing.T, cfg config.Config) *GithubScenarioV2 {
	return &GithubScenarioV2{
		id:       "replace-me",
		t:        t,
		client:   &GitHubClient{},
		actions:  make([]*actionV2, 0),
		reporter: NoopReporter{},
	}
}

func (s *GithubScenarioV2) Verbose() {
	s.reporter = &ConsoleReporter{}
}

func (s *GithubScenarioV2) Quiet() {
	s.reporter = NoopReporter{}
}

func (s *GithubScenarioV2) isApllied() bool {
	return s.appliedActionIdx >= len(s.actions)
}

func (s *GithubScenarioV2) Apply(ctx context.Context) error {
	s.t.Helper()
	s.reporter.Prefix("(Setup) ")
	defer s.reporter.Prefix("")
	var errs errors.MultiError
	setup := s.actions
	failFast := false

	for i, action := range setup {
		s.reporter.Writef("Applying '%s' = ", action.name)
		now := time.Now().UTC()

		var err error
		if i <= s.appliedActionIdx {
			err = action.apply(ctx)
			s.appliedActionIdx = i
		} else {
			s.reporter.Writeln("[SKIPPED]")
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
	return errs
}

// Apply implements Scenario.
func (s *GithubScenarioV2) Teardown(ctx context.Context) error {
	s.t.Helper()
	s.reporter.Prefix("(Teardown) ")
	defer s.reporter.Prefix("")
	var errs errors.MultiError
	teardown := reverse(s.actions)
	failFast := true

	for _, action := range teardown {
		s.reporter.Writef("Applying '%s' = ", action.name)
		now := time.Now().UTC()

		err := action.teardown(ctx)
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
	return errs
}

type Org struct {
	s    *GithubScenarioV2
	name string
}

func (o *Org) Get(ctx context.Context) (*github.Organization, error) {
	if o.s.isApllied() {
		return o.s.client.GetOrg(ctx, o.name)
	}
	panic("cannot retrieve org before scenario is applied")
}

func (o *Org) get(ctx context.Context) (*github.Organization, error) {
	return o.s.client.GetOrg(ctx, o.name)
}

func (o *Org) AllowPrivateForks(ctx context.Context) {
	updateOrgPermissions := &actionV2{
		name: "org:permissions:update:" + o.name,
		apply: func(ctx context.Context) error {

			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			org.MembersCanCreatePrivateRepos = boolp(true)
			org.MembersCanForkPrivateRepos = boolp(true)

			_, err = o.s.client.UpdateOrg(ctx, org)
			if err != nil {
				return err
			}
			return nil
		},
		teardown: func(ctx context.Context) error { return nil },
	}
	o.s.actions = append(o.s.actions, updateOrgPermissions)
}

type Teamv2 struct {
	s    *GithubScenarioV2
	org  *Org
	name string
}

func (team *Teamv2) Get(ctx context.Context) (*github.Team, error) {
	if team.s.isApllied() {
		return team.get(ctx)
	}
	panic("cannot retrieve org before scenario is applied")
}

func (team *Teamv2) get(ctx context.Context) (*github.Team, error) {
	return team.s.client.GetTeam(ctx, team.org.name, team.name)
}

func (tm *Teamv2) AddUser(u *User) {
	assignTeamMembership := &actionV2{
		name: "team:membership:%s:%s",
		apply: func(ctx context.Context) error {
			org, err := tm.org.get(ctx)
			if err != nil {
				return err
			}
			team, err := tm.get(ctx)
			if err != nil {
				return err
			}
			_, err = tm.s.client.assignTeamMembership(ctx, org, team, u.login)
			return err
		},
		teardown: func(ctx context.Context) error {
			return nil
		},
	}

	tm.s.actions = append(tm.s.actions, assignTeamMembership)
}

func (o *Org) CreateTeam(name string) *Teamv2 {
	baseTeam := &Teamv2{
		s: o.s,
	}

	createTeam := &actionV2{
		name: "org:team:create:" + name,
		apply: func(ctx context.Context) error {
			name := fmt.Sprintf("team-%s-%s", name, o.s.id)
			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			team, err := o.s.client.getOrCreateTeam(ctx, org, name)
			if err != nil {
				return err
			}
			baseTeam.name = team.GetName()
			return nil
		},
		teardown: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			return o.s.client.deleteTeam(ctx, org, baseTeam.name)
		},
	}

	o.s.actions = append(o.s.actions, createTeam)

	return baseTeam
}

func (s *GithubScenarioV2) CreateOrg(name string) *Org {
	baseOrg := &Org{}

	createOrg := &actionV2{
		name: "org:create:" + name,
		apply: func(ctx context.Context) error {
			orgName := fmt.Sprintf("org-%s-%s", name, s.id)
			org, err := s.client.createOrg(ctx, orgName)
			if err != nil {
				return err
			}
			baseOrg.name = org.GetName()
			return nil
		},
		teardown: func(context.Context) error {
			fmt.Printf("NEED TO MANUALLY DELETE: %s\n", baseOrg.name)
			return nil
		},
	}

	s.actions = append(s.actions, createOrg)
	return baseOrg
}

type User struct {
	s     *GithubScenarioV2
	name  string
	login string
}

func (s *GithubScenarioV2) CreateUser(name string) *User {
	baseUser := &User{}

	createUser := &actionV2{
		name: "user:create" + name,
		apply: func(ctx context.Context) error {
			name := fmt.Sprintf("user-%s-%s", name, s.id)
			email := "test-user-e2e@sourcegraph.com"
			user, err := s.client.createUser(ctx, name, email)
			if err != nil {
				return err
			}

			baseUser.name = user.GetName()
			baseUser.login = user.GetLogin()
			return nil
		},
		teardown: func(ctx context.Context) error {
			return s.client.deleteUser(ctx, baseUser.login)
		},
	}

	s.actions = append(s.actions, createUser)
	return baseUser
}
