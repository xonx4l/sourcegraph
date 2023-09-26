package tst

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
)

type Teamv2 struct {
	s    *GithubScenarioV2
	org  *Org
	name string
}

func (team *Teamv2) Get(ctx context.Context) (*github.Team, error) {
	if team.s.isApplied() {
		return team.get(ctx)
	}
	panic("cannot retrieve org before scenario is applied")
}

func (team *Teamv2) get(ctx context.Context) (*github.Team, error) {
	return team.s.client.GetTeam(ctx, team.org.name, team.name)
}

func (tm *Teamv2) AddUser(u *User) {
	assignTeamMembership := &actionV2{
		name: fmt.Sprintf("team:membership:%s:%s", tm.name, u.name),
		apply: func(ctx context.Context) error {
			org, err := tm.org.get(ctx)
			if err != nil {
				return err
			}
			team, err := tm.get(ctx)
			if err != nil {
				return err
			}
			user, err := u.get(ctx)
			if err != nil {
				return err
			}
			_, err = tm.s.client.assignTeamMembership(ctx, org, team, user)
			return err
		},
		teardown: nil,
	}

	tm.s.append(assignTeamMembership)
}
