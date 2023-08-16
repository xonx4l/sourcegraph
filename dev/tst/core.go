package tst

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ScenarioBuilder struct {
	ctx     context.Context
	client  *GithubClient
	store   scenarioStore
	actions []Action
}

type scenarioStore struct {
	store map[string]any
}

func (s *scenarioStore) SetOrg(org *github.Organization) {
	s.store["org"] = org
}
func (s *scenarioStore) GetOrg() (*github.Organization, error) {
	var result *github.Organization
	if v, ok := s.store["org"]; ok {
		if org, ok := v.(*github.Organization); ok {
			result = org
		} else {
			return nil, errors.New("failed to cast to github.Organization")
		}
	} else {
		return nil, errors.New("org not found - it might not have been loaded yet")
	}
	return result, nil
}

func (s *scenarioStore) SetScenarioUserMapping(u ScenarioUser, user *github.User) {
	s.store[string(u)] = user
}

func (s *scenarioStore) SetUsers(users []*github.User) {
	s.store["all-users"] = users
}

func (s *scenarioStore) GetUsers() ([]*github.User, error) {
	var result []*github.User
	if v, ok := s.store["org"]; ok {
		if org, ok := v.([]*github.User); ok {
			result = org
		} else {
			return nil, errors.New("failed to cast to []*github.User")
		}
	} else {
		return nil, errors.New("all-users not found - it might not have been loaded yet")
	}
	return result, nil
}

func (s *scenarioStore) GetScenarioUser(u ScenarioUser) (*github.User, error) {
	var result *github.User
	if v, ok := s.store[string(u)]; ok {
		if user, ok := v.(*github.User); ok {
			result = user
		} else {
			return nil, errors.New("failed to cast to github.User")
		}
	} else {
		return nil, errors.Newf("%s not found - it might not have been loaded yet", string(u))
	}
	return result, nil
}

func (s *scenarioStore) SetTeam(t *github.Team) {
	s.store[t.GetName()] = t
}

func (s *scenarioStore) GetTeam(name string) (*github.Team, error) {
	var result *github.Team
	if v, ok := s.store[name]; ok {
		if team, ok := v.(*github.Team); ok {
			result = team
		} else {
			return nil, errors.New("failed to cast to github.Team")
		}
	} else {
		return nil, errors.Newf("%s not found - it might not have been loaded yet", name)
	}
	return result, nil
}

type Kind string
type ScenarioUser string

var Github Kind = "Github"
var Bitbucket Kind = "Bitbucket"
var Gitlab Kind = "Gitlab"

var User1 ScenarioUser = "user1"
var User2 ScenarioUser = "user2"
var User3 ScenarioUser = "user3"
var User4 ScenarioUser = "user4"
var User5 ScenarioUser = "user5"
var User6 ScenarioUser = "user6"
var User7 ScenarioUser = "user7"
var User8 ScenarioUser = "user8"
var User9 ScenarioUser = "user9"
var User10 ScenarioUser = "user10"
var Admin ScenarioUser = "admin"

type Scenario interface {
}
type scenario struct{}

type ActionResult interface {
	Get() any
}

type ActionFn func(ctx context.Context) (ActionResult, error)

type Action interface {
	Name() string
	Hash() []byte
	Do(ctx context.Context) (ActionResult, error)
}

type action struct {
	name string
	hash []byte
	doFn ActionFn
}

func (a *action) Do(ctx context.Context) (ActionResult, error) {
	return a.doFn(ctx)
}

func (a *action) Hash() []byte {
	return a.hash
}

func (a *action) Name() string {
	return a.name
}

type GithubClient struct {
	c *github.Client
}

type actionResult[T any] struct {
	item T
}

func (a *actionResult[T]) Get() any {
	return a.item
}

func (gh *GithubClient) selectOrg(ctx context.Context) (*github.Organization, error) {
	org, resp, err := gh.c.Organizations.Get(ctx, "william-templates")
	if resp.StatusCode >= 299 {
		return nil, errors.Newf("failed to find org: %s", "william-templates")
	} else if err != nil {
		return nil, err
	}
	return org, err
}

func (gh *GithubClient) orgUsers(ctx context.Context, org *github.Organization) ([]*github.User, error) {
	users, _, err := gh.c.Organizations.ListMembers(ctx, org.GetName(), &github.ListMembersOptions{})
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (gh *GithubClient) getOrCreateTeam(ctx context.Context, org *github.Organization, name string) (*github.Team, error) {
	team, resp, err := gh.c.Teams.GetTeamBySlug(ctx, org.GetName(), name)

	switch resp.StatusCode {
	case 200:
		return team, nil
	case 404:
		newTeam := github.NewTeam{
			Name:        name,
			Description: strp("auto created team"),
			Privacy:     strp("closed"),
		}
		team, _, err = gh.c.Teams.CreateTeam(ctx, org.GetName(), newTeam)
	}
	return team, err
}

func (gh *GithubClient) assignTeamMembership(ctx context.Context, org *github.Organization, team *github.Team, users ...*github.User) (*github.Team, error) {
	for _, u := range users {
		_, resp, err := gh.c.Teams.GetTeamMembershipByID(ctx, org.GetID(), team.GetID(), u.GetLogin())
		if resp.StatusCode == 200 {
			// user is already part of this team
			return team, nil
		} else if resp.StatusCode >= 500 {
			return nil, errors.Newf("server error[%d]: %v", resp.StatusCode, err)
		}

		_, _, err = gh.c.Teams.AddTeamMembershipByID(ctx, org.GetID(), team.GetID(), u.GetLogin(), &github.TeamAddTeamMembershipOptions{
			Role: "member",
		})

		if err != nil {
			return nil, err
		}

	}
	return team, nil
}

func NewClient(kind Kind) *GithubClient {
	switch kind {
	case Github:
		return &GithubClient{}
	}
	return &GithubClient{}
}

type ScenarioTeam struct {
	name  string
	users []ScenarioUser
}

var _scenario Scenario = &scenario{}

func New(ctx context.Context, kind Kind) *ScenarioBuilder {
	return &ScenarioBuilder{ctx: ctx, client: NewClient(kind), actions: make([]Action, 0)}
}

func (sb *ScenarioBuilder) Org() *ScenarioBuilder {
	action := &action{
		name: "get-org",
		doFn: func(ctx context.Context) (ActionResult, error) {
			org, err := sb.client.selectOrg(ctx)
			if err != nil {
				return nil, err
			}

			sb.store.SetOrg(org)
			return &actionResult[*github.Organization]{item: org}, nil
		},
	}
	sb.actions = append(sb.actions, action)
	return sb
}

func (sb *ScenarioBuilder) Users(u ...ScenarioUser) *ScenarioBuilder {
	preloadAction := &action{
		name: "preload-users",
		doFn: func(ctx context.Context) (ActionResult, error) {
			org, err := sb.store.GetOrg()
			if err != nil {
				return nil, err
			}
			users, err := sb.client.orgUsers(ctx, org)
			if err != nil {
				return nil, err
			}
			sb.store.SetUsers(users)

			return &actionResult[[]*github.User]{item: users}, nil
		},
	}
	action := &action{
		name: "map-scenario-users",
		doFn: func(_ context.Context) (ActionResult, error) {
			users, err := sb.store.GetUsers()
			if err != nil {
				return nil, err
			}
			if len(u) > len(users) {
				return nil, errors.Newf("not enough users to use for scenario - required %d, available %d", len(u), len(users))
			}

			for i, user := range u {
				sb.store.SetScenarioUserMapping(user, users[i])
			}
			return &actionResult[bool]{item: true}, nil
		},
	}

	sb.actions = append(sb.actions, preloadAction, action)
	return sb
}

func Team(name string, u ...ScenarioUser) *ScenarioTeam {
	return &ScenarioTeam{
		name,
		u,
	}
}

func (sb *ScenarioBuilder) Teams(teams ...*ScenarioTeam) *ScenarioBuilder {
	for _, t := range teams {
		org, err := sb.store.GetOrg()
		createTeamAction := &action{
			name: "get-or-create-team:" + t.name,
			doFn: func(ctx context.Context) (ActionResult, error) {
				if err != nil {
					return nil, err
				}
				newTeam, err := sb.client.getOrCreateTeam(ctx, org, t.name)
				if err != nil {
					return nil, err
				}
				sb.store.SetTeam(newTeam)
				return &actionResult[*github.Team]{item: newTeam}, nil
			},
		}
		sb.actions = append(sb.actions, createTeamAction)

		assignTeamAction := &action{
			name: fmt.Sprintf("assign-team-membership:%s", t.name),
			doFn: func(ctx context.Context) (ActionResult, error) {
				team, err := sb.store.GetTeam(t.name)
				if err != nil {
					return nil, err
				}
				teamUsers := make([]*github.User, 0)
				for _, u := range t.users {
					if ghUser, err := sb.store.GetScenarioUser(u); err == nil {
						teamUsers = append(teamUsers, ghUser)
					} else {
						return nil, err
					}
					sb.client.assignTeamMembership(ctx, org, team, teamUsers...)
				}

				return &actionResult[*github.Team]{item: team}, nil
			},
		}
		sb.actions = append(sb.actions, assignTeamAction)
	}

	return sb
}

func (sb *ScenarioBuilder) Build() Scenario {
	for _, act := range sb.actions {
		fmt.Println(act.Name())
	}
	return scenario{}
}

func strp(v string) *string {
	return &v
}
