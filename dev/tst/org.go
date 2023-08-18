package tst

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
)

type GitHubScenarioOrg struct {
	ScenarioResource
}

func NewGitHubScenarioOrg(name string) *GitHubScenarioOrg {
	return &GitHubScenarioOrg{
		ScenarioResource: *NewScenarioResource(name),
	}
}

func (o *GitHubScenarioOrg) ID() string {
	return o.id
}

func (o *GitHubScenarioOrg) Name() string {
	return o.name
}

func (o *GitHubScenarioOrg) Key() string {
	return o.key
}

func (g GitHubScenarioOrg) CreateOrgAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		org, err := client.createOrg(ctx, g.Key())
		if err != nil {
			return nil, err
		}
		store.SetOrg(org)
		return &actionResult[*github.Organization]{item: org}, nil
	}

	return &action{
		name: fmt.Sprintf("create-org(%s)", g.Key()),
		doFn: fn,
	}
}

func (g GitHubScenarioOrg) DeleteOrgAction(client *GitHubClient) *action {
	fn := func(_ context.Context, store *scenarioStore) (ActionResult, error) {
		org, err := store.GetOrg()
		if err != nil {
			fmt.Printf("failed to find org: %v\n", err)
			return nil, err
		}
		fmt.Printf("NEED TO DELETE ORG: %s\n", org.GetLogin())
		return &actionResult[*github.Organization]{item: org}, nil
	}

	return &action{
		name: fmt.Sprintf("delete-org(%s)", g.Key()),
		doFn: fn,
	}
}
