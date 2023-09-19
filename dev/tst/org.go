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

func (g GitHubScenarioOrg) CreateOrgAction(client *GitHubClient) Action {
	return &action{
		id:   g.Key(),
		name: "create-org",
		fn: func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
			org, err := client.createOrg(ctx, g.Key())
			if err != nil {
				return nil, err
			}
			store.SetOrg(org)
			return &actionResult[*github.Organization]{item: org}, nil
		},
	}
}

func (g GitHubScenarioOrg) UpdateOrgPermissionsAction(client *GitHubClient) Action {
	return &action{
		id:   g.Key(),
		name: "update-org-permissions",
		fn: func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}

			org.MembersCanCreatePrivateRepos = boolp(true)
			org.MembersCanForkPrivateRepos = boolp(true)

			org, err = client.UpdateOrg(ctx, org)
			if err != nil {
				return nil, err
			}
			store.SetOrg(org)
			return &actionResult[*github.Organization]{item: org}, nil
		},
	}

}

func (g GitHubScenarioOrg) DeleteOrgAction(client *GitHubClient) Action {
	return &action{
		id:   g.Key(),
		name: fmt.Sprintf("delete-org(%s)", g.Key()),
		fn: func(_ context.Context, store *scenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}
			fmt.Printf("NEED TO DELETE ORG: %s\n", org.GetLogin())
			return &actionResult[*github.Organization]{item: org}, nil
		},
	}
}
