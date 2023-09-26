package tst

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
)

type User struct {
	s    *GithubScenarioV2
	name string
}

func (u *User) Get(ctx context.Context) (*github.User, error) {
	if u.s.IsApplied() {
		return u.get(ctx)
	}
	panic("cannot retrieve user before scenario is applied")
}

func (u *User) get(ctx context.Context) (*github.User, error) {
	return u.s.client.getUser(ctx, u.name)
}

func (s *GithubScenarioV2) CreateUser(name string) *User {
	baseUser := &User{
		s:    s,
		name: name,
	}

	createUser := &actionV2{
		name: "user:create:" + name,
		apply: func(ctx context.Context) error {
			name := fmt.Sprintf("user-%s-%s", name, s.id)
			email := "test-user-e2e@sourcegraph.com"
			user, err := s.client.createUser(ctx, name, email)
			if err != nil {
				return err
			}

			baseUser.name = user.GetLogin()
			return nil
		},
		teardown: func(ctx context.Context) error {
			return s.client.deleteUser(ctx, baseUser.name)
		},
	}

	s.append(createUser)
	return baseUser
}
