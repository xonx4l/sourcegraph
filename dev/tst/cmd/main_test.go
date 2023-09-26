package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/dev/tst"
	"github.com/sourcegraph/sourcegraph/dev/tst/config"
)

// PLAN
// Requirements
// Running Sourcegraph
// new Org
// * 2 Repos
// ** 1 repo with private repo
// * 1 user

func TestRepo(t *testing.T) {
	cfg, err := config.FromFile("config.json")
	if err != nil {
		t.Fatalf("error loading scenario config: %v\n", err)
	}
	scenario, err := tst.NewGithubScenarioV2(context.Background(), t, *cfg)
	if err != nil {
		t.Fatalf("error creating scenario: %v\n", err)
	}
	// s := builder.Org("tst-org").
	// 	Users(tst.Admin, tst.User1).
	// 	Teams(tst.Team("public-team", tst.Admin), tst.Team("private-team", tst.User1)).
	// 	Repos(tst.PublicRepo("sgtest/go-diff", "public-team", true), tst.PrivateRepo("sgtest/private", "private-team", true))
	org := scenario.CreateOrg("tst-org")
	user := scenario.CreateUser("tst-user")

	//ctx := context.Background()

	org.AllowPrivateForks()
	team := org.CreateTeam("team-1")
	team.AddUser(user)

	fmt.Println(scenario.Plan())

	fmt.Println()
	fmt.Printf("Applying scenario")
	scenario.Verbose()
	if err := scenario.Apply(context.Background()); err != nil {
		t.Fatalf("error applying scenario: %v", err)
	}

}
