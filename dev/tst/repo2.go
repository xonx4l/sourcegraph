package tst

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Repov2 struct {
	s    *GithubScenarioV2
	team *Teamv2
	org  *Org
	name string
}

func (r *Repov2) Get(ctx context.Context) (*github.Repository, error) {
	return nil, nil
}

func (r *Repov2) get(ctx context.Context) (*github.Repository, error) {
	return nil, nil
}

func (r *Repov2) Fork(target string) {
	action := &actionV2{
		name: fmt.Sprintf("repo:fork:%s", target),
		apply: func(ctx context.Context) error {
			org, err := r.org.get(ctx)
			if err != nil {
				return err
			}

			var owner, repoName string
			parts := strings.Split(target, "/")
			if len(parts) >= 2 {
				owner = parts[0]
				repoName = parts[1]
			} else {
				return errors.Newf("incorrect repo format for %q - expecting {owner}/{name}")
			}

			err = r.s.client.forkRepo(ctx, org, owner, repoName)
			if err != nil {
				return err
			}

			// Wait till fork has synced
			time.Sleep(1 * time.Second)
			r.name = fmt.Sprintf(org.GetLogin(), repoName)
			return nil
		},
		teardown: func(ctx context.Context) error {
			repo, err := r.get(ctx)
			if err != nil {
				return err
			}

			org, err := r.org.get(ctx)
			if err != nil {
				return err
			}

			return r.s.client.deleteRepo(ctx, org, repo)
		},
	}

	r.s.append(action)
}

func (r *Repov2) Create(public bool) {
	action := &actionV2{
		name: fmt.Sprintf("repo:create:%s", r.name),
		apply: func(ctx context.Context) error {
			org, err := r.org.get(ctx)
			if err != nil {
				return err
			}

			var repoName string
			parts := strings.Split(r.name, "/")
			if len(parts) >= 2 {
				repoName = parts[1]
			} else {
				return errors.Newf("incorrect repo format for %q - expecting {owner}/{name}")
			}

			repo, err := r.s.client.newRepo(ctx, org, repoName, public)
			if err != nil {
				return err
			}

			r.name = repo.GetFullName()
			return nil
		},
		teardown: func(ctx context.Context) error {
			repo, err := r.get(ctx)
			if err != nil {
				return err
			}

			org, err := r.org.get(ctx)
			if err != nil {
				return err
			}

			return r.s.client.deleteRepo(ctx, org, repo)
		},
	}

	r.s.append(action)
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
