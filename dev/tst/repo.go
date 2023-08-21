package tst

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubScenarioRepo struct {
	ScenarioResource
	teamName string
	fork     bool
	private  bool
}

func NewGitHubScenarioRepo(name, teamKey string, fork, private bool) *GitHubScenarioRepo {
	return &GitHubScenarioRepo{
		ScenarioResource: *NewScenarioResource(name),
		teamName:         teamKey,
		fork:             fork,
		private:          private,
	}
}

func (gr *GitHubScenarioRepo) ForkRepoAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}

		var owner, repoName string
		parts := strings.Split(gr.name, "/")
		if len(parts) >= 2 {
			owner = parts[0]
			repoName = parts[1]
		} else {
			return nil, errors.Newf("incorrect repo format for %q - expecting {owner}/{name}")
		}

		err = client.forkRepo(ctx, org, owner, repoName)
		if err != nil {
			return nil, err
		}
		return &actionResult[bool]{item: true}, nil
	}

	return &action{
		name: fmt.Sprintf("fork-repo(%s)", gr.Key()),
		doFn: fn,
	}
}

func (gr *GitHubScenarioRepo) GetForkedRepo(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		// Wait till fork has synced
		time.Sleep(1 * time.Second)
		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}

		var repoName string
		parts := strings.Split(gr.name, "/")
		if len(parts) >= 2 {
			repoName = parts[1]
		} else {
			repoName = parts[0]
		}
		if gr.fork && repoName == "" {
			return nil, errors.Newf("incorrect repo format for %q - expecting {owner}/{name}")
		}

		repo, err := client.getRepo(ctx, org.GetLogin(), repoName)
		if err != nil {
			return nil, err
		}
		// Since this is a forked repo we need to update the GitHubScenarioRepo
		// We only edit the name but the id stays the same because the initial name
		// is "someorg/repo" and the name should reflect the name with the current org
		// "currentOrg/repo"
		// TODO: this is nasty - find a better way
		gr.name = repo.GetFullName()
		store.SetRepo(gr, repo)
		return &actionResult[bool]{item: true}, nil
	}

	return &action{
		name: fmt.Sprintf("get-fork-repo(%s)", gr.Key()),
		doFn: fn,
	}
}

func (gr *GitHubScenarioRepo) InitRepoAction(client *GitHubClient) *action {
	return &action{
		name: fmt.Sprintf("init-repo(&s)", gr.Key()),
		doFn: nil,
	}
}

func (gr *GitHubScenarioRepo) PushRepoAction(client *GitHubClient) *action {
	return &action{
		name: fmt.Sprintf("push-repo(&s)", gr.Key()),
		doFn: nil,
	}
}

func (gr *GitHubScenarioRepo) NewRepo(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}

		var repoName string
		parts := strings.Split(gr.name, "/")
		if len(parts) >= 2 {
			repoName = parts[1]
		} else {
			repoName = parts[0]
		}

		repo, err := client.newRepo(ctx, org, repoName, gr.private)
		if err != nil {
			return nil, err
		}

		gr.name = repo.GetFullName()
		store.SetRepo(gr, repo)

		return &actionResult[bool]{item: true}, nil
	}

	return &action{
		name: fmt.Sprintf("create-repo(%s)", gr.Key()),
		doFn: fn,
	}
}

func (gr *GitHubScenarioRepo) SetPermissionsAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		repo, err := store.GetRepo(gr)
		if err != nil {
			return nil, err
		}

		repo.Private = &gr.private

		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}

		repo, err = client.updateRepo(ctx, org, repo)
		if err != nil {
			return nil, err
		}
		store.SetRepo(gr, repo)
		return &actionResult[*github.Repository]{item: repo}, nil
	}
	return &action{
		name: fmt.Sprintf("repo-permissions(%s)", gr.Key()),
		doFn: fn,
	}
}

func (gr *GitHubScenarioRepo) DeleteRepoAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		repo, err := store.GetRepo(gr)
		if err != nil {
			return nil, err
		}

		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}

		err = client.deleteRepo(ctx, org, repo)
		if err != nil {
			return nil, err
		}
		store.SetRepo(gr, repo)
		return &actionResult[bool]{item: true}, nil
	}
	return &action{
		name: fmt.Sprintf("delete-repo(%s)", gr.Key()),
		doFn: fn,
	}
}

func (gr *GitHubScenarioRepo) AssignTeamAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}

		repo, err := store.GetRepo(gr)
		if err != nil {
			return nil, err
		}

		team, err := store.GetTeamByName(gr.teamName)
		if err != nil {
			return nil, err
		}

		err = client.updateTeamRepoPermissions(ctx, org, team, repo)
		if err != nil {
			return nil, err
		}
		store.SetRepo(gr, repo)
		return &actionResult[bool]{item: true}, nil
	}
	return &action{
		name: fmt.Sprintf("assign-team(%s)", gr.key),
		doFn: fn,
	}
}
