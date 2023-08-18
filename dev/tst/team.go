package tst

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
)

type GitHubScenarioTeam struct {
	name  string
	id    string
	key   string
	users []GitHubScenarioUser
}

func NewGitHubScenarioTeam(name string, u ...GitHubScenarioUser) *GitHubScenarioTeam {
	id := id()
	key := joinID(name, "-", id, 39)
	return &GitHubScenarioTeam{
		name:  name,
		id:    id,
		key:   key,
		users: []GitHubScenarioUser{},
	}
}

func (t *GitHubScenarioTeam) ID() string {
	return t.id
}

func (t *GitHubScenarioTeam) Name() string {
	return t.name
}

func (t *GitHubScenarioTeam) Key() string {
	return t.key
}

func (gt *GitHubScenarioTeam) CreateTeamAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}
		newTeam, err := client.getOrCreateTeam(ctx, org, gt.name)
		if err != nil {
			return nil, err
		}
		store.SetTeam(gt, newTeam)
		return &actionResult[*github.Team]{item: newTeam}, nil
	}

	return &action{
		name: fmt.Sprintf("get-or-create-team(%s)", gt.name),
		doFn: fn,
	}
}

func (gt *GitHubScenarioTeam) DeleteTeamAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}
		err = client.deleteTeam(ctx, org, gt.name)
		if err != nil {
			return nil, err
		}
		return &actionResult[bool]{item: true}, nil
	}

	return &action{
		name: fmt.Sprintf("delete-team(%s)", gt.name),
		doFn: fn,
	}
}

func (gt *GitHubScenarioTeam) AssignTeamAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}
		team, err := store.GetTeam(gt)
		if err != nil {
			return nil, err
		}
		teamUsers := make([]*github.User, 0)
		for _, u := range gt.users {
			if ghUser, err := store.GetScenarioUser(u); err == nil {
				teamUsers = append(teamUsers, ghUser)
			} else {
				return nil, err
			}
			client.assignTeamMembership(ctx, org, team, teamUsers...)
		}

		return &actionResult[*github.Team]{item: team}, nil
	}

	return &action{
		name: fmt.Sprintf("assign-team-membership(%s)", gt.Key()),
		doFn: fn,
	}
}
