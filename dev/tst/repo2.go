package tst

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
)

type Repov2 struct {
	s    *GithubScenarioV2
	team *Teamv2
	org  *Org
	name string
}

func (r *Repov2) Get(ctx context.Context) (*github.Repository, error) {
	if r.s.IsApplied() {
		return r.get(ctx)
	}
	panic("cannot retrieve repo before scenario is applied")
}

func (r *Repov2) get(ctx context.Context) (*github.Repository, error) {
	return nil, nil
}

func (r *Repov2) AddTeam(team *Teamv2) {
	r.team = team
	action := &actionV2{
		name: fmt.Sprintf("repo:team:%s:membership:%s", team.name, r.name),
		apply: func(ctx context.Context) error {
			org, err := r.org.get(ctx)
			if err != nil {
				return err
			}

			repo, err := r.get(ctx)
			if err != nil {
				return err
			}

			team, err := r.team.get(ctx)
			if err != nil {
				return err
			}

			err = r.s.client.updateTeamRepoPermissions(ctx, org, team, repo)
			if err != nil {
				return err
			}
			return nil
		},
		teardown: nil,
	}

	r.s.append(action)
}

func (r *Repov2) SetPermissions(private bool) {
	permissionKey := "private"
	if !private {
		permissionKey = "public"
	}
	action := &actionV2{
		name: fmt.Sprintf("repo:permissions:%s:%s", r.name, permissionKey),
		apply: func(ctx context.Context) error {
			repo, err := r.get(ctx)
			if err != nil {
				return err
			}
			repo.Private = &private

			org, err := r.org.get(ctx)
			if err != nil {
				return err
			}

			repo, err = r.s.client.updateRepo(ctx, org, repo)
			if err != nil {
				return err
			}
			return err
		},
	}

	r.s.append(action)
}
