package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/dev/tst"
)

// PLAN
// Requirements
// Running Sourcegraph
// new Org
// * 2 Repos
// ** 1 repo with private repo
// * 1 user
var tstCfg *tst.Config

func TestRepo(t *testing.T) {
	builder, err := tst.NewGitHubScenario(context.Background(), tstCfg, t)
	if err != nil {
		fmt.Printf("failed to create scenario: %v", err)
	}

	s := builder.Org("tst-org").
		Users(tst.Admin, tst.User1).
		Teams(tst.Team("public-team", tst.Admin), tst.Team("private-team", tst.User1)).
		Repos(tst.PublicRepo("sgtest/go-diff", "public-team", true), tst.PrivateRepo("sgtest/private", "private-team", true))

	fmt.Println(s)

	ctx := context.Background()
	_, teardown, err := s.Setup(ctx)
	if err != nil {
		fmt.Printf("error during scenario setup: %v\n", err)
		teardown(ctx)
		os.Exit(1)
	}
	defer teardown(ctx)
}

func main() {
	cfg, err := tst.LoadConfig("config.json")
	if err != nil {
		fmt.Printf("error loading scenario config: %v\n", err)
	}
	tstCfg = cfg
}
