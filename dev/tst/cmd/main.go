package main

import (
	"context"

	"github.com/sourcegraph/sourcegraph/dev/tst"
)

func main() {
	scenario := tst.New(context.Background(), tst.Github).
		Org().
		Users(tst.User1, tst.User2, tst.User3, tst.Admin).
		Teams(tst.Team("team1", tst.User1), tst.Team("team2", tst.User2), tst.Team("team3", tst.User3, tst.Admin)).
		Repos(
			tst.PublicRepo("repo/public"),
			tst.PrivateRepo("repo/pvt1", "team1"),
			tst.PrivateRepo("repo/pvt2", "team2"),
			tst.PrivateRepo("repo/pvt3", "team3"),
		)
	scenario.Build()
}
