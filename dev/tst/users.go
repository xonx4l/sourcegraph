package tst

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubScenarioUser struct {
	ScenarioResource
}

func NewGitHubScenarioUser(name string) *GitHubScenarioUser {
	return &GitHubScenarioUser{
		ScenarioResource: *NewScenarioResource(name),
	}
}

func (u *GitHubScenarioUser) ID() string {
	return u.id
}

func (u *GitHubScenarioUser) Name() string {
	return u.name
}

func (u *GitHubScenarioUser) Key() string {
	if u == &Admin {
		return "admin"
	}
	return u.key
}

var User1 GitHubScenarioUser = *NewGitHubScenarioUser("user1")
var User2 GitHubScenarioUser = *NewGitHubScenarioUser("user2")
var User3 GitHubScenarioUser = *NewGitHubScenarioUser("user3")
var User4 GitHubScenarioUser = *NewGitHubScenarioUser("user4")
var User5 GitHubScenarioUser = *NewGitHubScenarioUser("user5")
var User6 GitHubScenarioUser = *NewGitHubScenarioUser("user6")
var User7 GitHubScenarioUser = *NewGitHubScenarioUser("user7")
var User8 GitHubScenarioUser = *NewGitHubScenarioUser("user8")
var User9 GitHubScenarioUser = *NewGitHubScenarioUser("user9")
var User10 GitHubScenarioUser = *NewGitHubScenarioUser("user10")
var Admin GitHubScenarioUser = *NewGitHubScenarioUser("admin")

func preloadUsersAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		org, err := store.GetOrg()
		if err != nil {
			return nil, err
		}
		users, err := client.orgUsers(ctx, org)
		if err != nil {
			return nil, err
		}
		store.SetUsers(users)

		return &actionResult[[]*github.User]{item: users}, nil
	}
	return &action{
		name: "preload-users()",
		doFn: fn,
	}
}

func mapUsersAction(_ *GitHubClient, scenarioUsers []*GitHubScenarioUser) *action {
	fn := func(_ context.Context, store *scenarioStore) (ActionResult, error) {
		users, err := store.GetUsers()
		if err != nil {
			return nil, err
		}
		if len(scenarioUsers) > len(users) {
			return nil, errors.Newf("not enough users to use for scenario - required %d, available %d", len(scenarioUsers), len(users))
		}

		for i, user := range scenarioUsers {
			store.SetScenarioUserMapping(user, users[i])
		}
		return &actionResult[bool]{item: true}, nil
	}

	return &action{
		name: "map-scenario-users()",
		doFn: fn,
	}
}

func userEmail(u *GitHubScenarioUser) string {
	return "william.bezuidenhout@sourcegraph.com" //fmt.Sprintf("william.bezuidenhout@sourcegraph.com", u.Key())
}

func (u *GitHubScenarioUser) GetUserAction(client *GitHubClient) *action {
	name := u.Key()
	if u.Name() == Admin.Name() {
		name = client.cfg.User
	}
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		user, err := client.getUser(ctx, name)
		if err != nil {
			return nil, err
		}

		store.SetScenarioUserMapping(u, user)
		return &actionResult[*github.User]{item: user}, nil
	}

	return &action{
		name: fmt.Sprintf("get-user(%s)", name),
		doFn: fn,
	}
}

func (u *GitHubScenarioUser) CreateUserAction(client *GitHubClient) *action {
	name := u.Key()
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		user, err := client.createUser(ctx, "test", userEmail(u))
		if err != nil {
			return nil, err
		}

		store.SetScenarioUserMapping(u, user)
		return &actionResult[*github.User]{item: user}, nil
	}

	return &action{
		name: fmt.Sprintf("create-user(%s)", name),
		doFn: fn,
	}
}

func (u GitHubScenarioUser) DeleteUserAction(client *GitHubClient) *action {
	fn := func(ctx context.Context, store *scenarioStore) (ActionResult, error) {
		user, err := store.GetScenarioUser(u)
		if err != nil {
			return nil, err
		}
		err = client.deleteUser(ctx, user.GetLogin())
		if err != nil {
			return nil, err
		}

		return &actionResult[*github.User]{item: user}, nil
	}

	return &action{
		name: fmt.Sprintf("delete-user(%s)", u.Key()),
		doFn: fn,
	}
}
