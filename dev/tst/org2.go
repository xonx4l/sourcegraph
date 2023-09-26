package tst

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
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
}
