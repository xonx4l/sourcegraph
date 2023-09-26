package tst

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Org struct {
	s    *GithubScenarioV2
	name string
}

func (o *Org) Get(ctx context.Context) (*github.Organization, error) {
	if o.s.IsApplied() {
		return o.s.client.GetOrg(ctx, o.name)
	}
	panic("cannot retrieve org before scenario is applied")
}

func (o *Org) get(ctx context.Context) (*github.Organization, error) {
	return o.s.client.GetOrg(ctx, o.name)
}

func (o *Org) AllowPrivateForks() {
	updateOrgPermissions := &actionV2{
		name: "org:permissions:update:" + o.name,
		apply: func(ctx context.Context) error {

			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			org.MembersCanCreatePrivateRepos = boolp(true)
			org.MembersCanForkPrivateRepos = boolp(true)

			_, err = o.s.client.UpdateOrg(ctx, org)
			if err != nil {
				return err
			}
			return nil
		},
		teardown: nil,
	}
	o.s.append(updateOrgPermissions)
}

func (o *Org) CreateTeam(name string) *Teamv2 {
	baseTeam := &Teamv2{
		s:    o.s,
		org:  o,
		name: name,
	}

	createTeam := &actionV2{
		name: "org:team:create:" + name,
		apply: func(ctx context.Context) error {
			name := fmt.Sprintf("team-%s-%s", name, o.s.id)
			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			team, err := o.s.client.getOrCreateTeam(ctx, org, name)
			if err != nil {
				return err
			}
			baseTeam.name = team.GetName()
			return nil
		},
		teardown: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			return o.s.client.deleteTeam(ctx, org, baseTeam.name)
		},
	}

	o.s.append(createTeam)

	return baseTeam
}

func (o *Org) CreateRepo(name string, public bool) *Repov2 {
	baseRepo := &Repov2{
		s:    o.s,
		org:  o,
		name: name,
	}
	action := &actionV2{
		name: fmt.Sprintf("repo:create:%s", name),
		apply: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}

			var repoName string
			parts := strings.Split(name, "/")
			if len(parts) >= 2 {
				repoName = parts[1]
			} else {
				return errors.Newf("incorrect repo format for %q - expecting {owner}/{name}")
			}

			repo, err := o.s.client.newRepo(ctx, org, repoName, public)
			if err != nil {
				return err
			}

			baseRepo.name = repo.GetFullName()
			return nil
		},
		teardown: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}

			repo, err := baseRepo.get(ctx)
			if err != nil {
				return err
			}

			return o.s.client.deleteRepo(ctx, org, repo)
		},
	}
	o.s.append(action)

	return baseRepo
}

func (o *Org) CreateRepoFork(target string) *Repov2 {
	baseRepo := &Repov2{
		s:    o.s,
		org:  o,
		name: target,
	}
	action := &actionV2{
		name: fmt.Sprintf("repo:fork:%s", target),
		apply: func(ctx context.Context) error {
			org, err := o.get(ctx)
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

			err = o.s.client.forkRepo(ctx, org, owner, repoName)
			if err != nil {
				return err
			}

			// Wait till fork has synced
			time.Sleep(1 * time.Second)
			baseRepo.name = fmt.Sprintf(org.GetLogin(), repoName)
			return nil
		},
		teardown: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}

			repo, err := baseRepo.get(ctx)
			if err != nil {
				return err
			}

			return o.s.client.deleteRepo(ctx, org, repo)
		},
	}
	o.s.append(action)

	return baseRepo
}
