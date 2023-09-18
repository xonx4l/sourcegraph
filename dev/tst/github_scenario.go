package tst

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/v53/github"
)

type ScenarioResource struct {
	name string
	id   string
	key  string
}

func NewScenarioResource(name string) *ScenarioResource {
	id := id()
	key := joinID(name, "-", id, 39)
	return &ScenarioResource{
		name: name,
		id:   id,
		key:  key,
	}

}

func (s *ScenarioResource) ID() string {
	return s.id
}

func (s *ScenarioResource) Name() string {
	return s.name
}

func (s *ScenarioResource) Key() string {
	return s.key
}

type GitHubScenarioBuilder struct {
	test    *testing.T
	client  *GitHubClient
	store   *scenarioStore
	actions *actionManager
}

type Scenario interface {
}

type scenario struct {
	client *GitHubClient
	users  []*github.User
	teams  []*github.Team
	repos  []*github.Repository
	org    *github.Organization
}

var _scenario Scenario = &scenario{}

func NewGitHubScenario(ctx context.Context, t *testing.T, cfg *Config) (*GitHubScenarioBuilder, error) {
	client, err := NewGitHubClient(ctx, *cfg)
	if err != nil {
		return nil, err
	}

	return &GitHubScenarioBuilder{
		test:    t,
		client:  client,
		store:   NewStore(),
		actions: NewActionManager(),
	}, nil
}

func (sb *GitHubScenarioBuilder) T(t *testing.T) *GitHubScenarioBuilder {
	sb.test = t
	return sb
}

func (sb *GitHubScenarioBuilder) Org(name string) *GitHubScenarioBuilder {
	sb.test.Helper()
	org := NewGitHubScenarioOrg(name)
	sb.actions.AddSetup(org.CreateOrgAction(sb.client), org.UpdateOrgPermissionsAction(sb.client))
	sb.actions.AddTeardown(org.DeleteOrgAction(sb.client))
	return sb
}

func (sb *GitHubScenarioBuilder) Users(users ...GitHubScenarioUser) *GitHubScenarioBuilder {
	sb.test.Helper()
	for _, u := range users {
		if u == Admin {
			sb.actions.AddSetup(u.GetUserAction(sb.client))
		} else {
			sb.actions.AddSetup(u.CreateUserAction(sb.client))
			sb.actions.AddTeardown(u.DeleteUserAction(sb.client))
		}
	}
	return sb
}

func Team(name string, u ...GitHubScenarioUser) *GitHubScenarioTeam {
	return NewGitHubScenarioTeam(name, u...)
}

func (sb *GitHubScenarioBuilder) Teams(teams ...*GitHubScenarioTeam) *GitHubScenarioBuilder {
	sb.test.Helper()
	for _, t := range teams {
		sb.actions.AddSetup(t.CreateTeamAction(sb.client), t.AssignTeamAction(sb.client))
		sb.actions.AddTeardown(t.DeleteTeamAction(sb.client))
	}

	return sb
}

func (sb *GitHubScenarioBuilder) Repos(repos ...*GitHubScenarioRepo) *GitHubScenarioBuilder {
	sb.test.Helper()
	for _, r := range repos {
		if r.fork {
			sb.actions.AddSetup(r.ForkRepoAction(sb.client), r.GetRepoAction(sb.client))
			// Seems like you can't change permissions for a repo fork
			//sb.setupActions = append(sb.setupActions, r.SetPermissionsAction(sb.client))
			sb.actions.AddTeardown(r.DeleteRepoAction(sb.client))
		} else {
			sb.actions.AddSetup(r.NewRepoAction(sb.client),
				r.GetRepoAction(sb.client),
				r.InitLocalRepoAction(sb.client),
				r.SetPermissionsAction(sb.client),
			)

			sb.actions.AddTeardown(r.DeleteRepoAction(sb.client))
		}
		sb.actions.AddSetup(r.AssignTeamAction(sb.client))
	}

	return sb
}

func PublicRepo(name string, team string, fork bool) *GitHubScenarioRepo {
	return NewGitHubScenarioRepo(name, team, fork, false)
}

func PrivateRepo(name string, team string, fork bool) *GitHubScenarioRepo {
	return NewGitHubScenarioRepo(name, team, fork, true)
}

func (sb *GitHubScenarioBuilder) Setup(ctx context.Context) (Scenario, func(context.Context) error, error) {
	sb.test.Helper()
	fmt.Println("-- Setup --")
	start := time.Now().UTC()
	err := sb.actions.Apply(ctx, &actionApplyCfg{
		test:     sb.test,
		store:    sb.store,
		actions:  sb.actions.setup,
		failFast: false,
	})
	fmt.Printf("Run complete: %s\n", time.Now().UTC().Sub(start))
	return scenario{}, sb.TearDown, err
}

func (sb *GitHubScenarioBuilder) TearDown(ctx context.Context) error {
	sb.test.Helper()
	fmt.Println("-- Teardown --")
	start := time.Now().UTC()
	err := sb.actions.Apply(ctx, &actionApplyCfg{
		test:     sb.test,
		store:    sb.store,
		actions:  reverse(sb.actions.teardown),
		failFast: false,
	})
	fmt.Printf("Run complete: %s\n", time.Now().UTC().Sub(start))
	fmt.Println("-- Teardown --")
	return err
}

func (sb *GitHubScenarioBuilder) String() string {
	return sb.actions.String()
}
