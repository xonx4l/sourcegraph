package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"

	"github.com/shurcooL/graphql"
)

type CodeHost struct {
	Kind     string `json:"Kind"`
	Token    string `json:"Token"`
	Org      string `json:"Org"`
	URL      string `json:"URL"`
	User     string `json:"User"`
	Password string `json:"Password"`
}

type SourcegraphCfg struct {
	URL   string `json:"URL"`
	User  string `json:"User"`
	Token string `json:"Token"`
}

type config struct {
	CodeHost    CodeHost       `json:"CodeHost"`
	Sourcegraph SourcegraphCfg `json:"Sourcegraph"`
}

func Load(filename string) (*config, error) {
	var c config

	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(fd).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

type Client struct {
	config *config
	org    *github.Organization
	gh     *github.Client
}

func NewClient(ctx context.Context, cfg config) (*Client, error) {
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.CodeHost.Token},
	))

	if true {
		tc.Transport.(*oauth2.Transport).Base = http.DefaultTransport
		tc.Transport.(*oauth2.Transport).Base.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	gh, err := github.NewEnterpriseClient(cfg.CodeHost.URL, cfg.CodeHost.URL, tc)
	if err != nil {
		log.Fatalf("failed to creaTe enterprise cLient: %v\n", err)
	}

	org, _, err := gh.Organizations.Get(ctx, cfg.CodeHost.Org)
	if err != nil {
		return nil, err
	}

	c := Client{
		gh:     gh,
		org:    org,
		config: &cfg,
	}

	return &c, err
}

func (c *Client) AddTeamMembership(ctx context.Context, user *github.User, team *github.Team) error {
	// is this user already part of the team?
	_, resp, err := c.gh.Teams.GetTeamMembershipByID(ctx, c.org.GetID(), team.GetID(), user.GetLogin())
	if resp.StatusCode == 200 {
		// user is already part of this team
		return nil
	} else if resp.StatusCode >= 500 {
		return fmt.Errorf("server error[%d]: %v", resp.StatusCode, err)
	}

	// user isn't part of the team so lets add them
	log.Printf("[INFO] Add user %q to team %s", user.GetLogin(), team.GetName())
	_, _, err = c.gh.Teams.AddTeamMembershipByID(ctx, c.org.GetID(), team.GetID(), user.GetLogin(), &github.TeamAddTeamMembershipOptions{
		Role: "member",
	})

	return err

}

func (c *Client) GetOrCreateTeam(ctx context.Context, newTeam *github.NewTeam) (*github.Team, error) {
	team, resp, err := c.gh.Teams.GetTeamBySlug(ctx, c.config.CodeHost.Org, newTeam.Name)
	switch resp.StatusCode {
	case 200:
		return team, nil
	case 404:
		team, _, err = c.gh.Teams.CreateTeam(ctx, c.config.CodeHost.Org, *newTeam)
	}
	return team, err
}

type TemplateUser struct {
	UserKey string
	User    *github.User
	Teams   []*github.Team
}

func NewTemplateUser(userKey string) *TemplateUser {
	return &TemplateUser{
		UserKey: userKey,
		User:    nil,
		Teams:   make([]*github.Team, 0),
	}
}

func (t *TemplateUser) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Name: %s (%s)\n", t.UserKey, t.User.GetLogin()))
	sb.WriteString("Teams\n")
	for _, tt := range t.Teams {
		sb.WriteString(fmt.Sprintf("- %s [%d]\n", tt.GetName(), tt.GetReposCount()))
	}

	return sb.String()
}

var userMap = map[string]string{
	"indradhanush":     "user1",
	"integration-test": "user2",
	"milton":           "admin",
	"testing":          "user3",
}

func (c *Client) GetOrCreateRepo(ctx context.Context, team *github.Team, repo *github.Repository) (*github.Repository, error) {
	currRepo, resp, err := c.gh.Repositories.Get(ctx, c.config.CodeHost.Org, repo.GetName())

	switch resp.StatusCode {
	case 200:
		log.Printf("[INFO] repo %s already exists - returning\n", repo.GetName())
		return currRepo, nil
	case 422:
		return nil, err
	case 500:
		return nil, err
	}
	r, _, err := c.gh.Repositories.Create(ctx, c.config.CodeHost.Org, repo)
	log.Printf("[INFO] created repo %s\n", repo.GetName())

	return r, err
}

func (c *Client) OrgUsers(ctx context.Context) ([]*github.User, error) {
	users, _, err := c.gh.Organizations.ListMembers(ctx, c.config.CodeHost.Org, &github.ListMembersOptions{})
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Scenario
// Users:
// - Admin
// - User1
// - User2
//
// Repos:
// - repo/public
// - Admin
// - User1
// - User2
// - repo/private1
// - User1
// - repo/private2
// - User2
// - repo/private3
// - User1
// - User2

func strp(v string) *string {
	return &v
}

func (c *Client) Cleanup(ctx context.Context, templateUsers map[string]*TemplateUser, teamRepoMap map[string]*github.Repository) error {

	if c.config.CodeHost.Org == "milton" {
		return fmt.Errorf("org is milton - not cleaning up")
	}

	// Loop through each template user
	deletedTeams := make(map[int64]bool, 0)
	for _, tu := range templateUsers {

		// Delete the user's repositories
		if repo, ok := teamRepoMap[tu.UserKey]; ok {
			_, err := c.gh.Repositories.Delete(ctx, c.config.CodeHost.Org, repo.GetName())
			if err != nil {
				return err
			}
			log.Printf("[INFO] Deleted repo %s", repo.GetName())
		}

		// Delete the user's teams
		for _, team := range tu.Teams {
			if _, deleted := deletedTeams[team.GetID()]; deleted {
				log.Printf("[DEBUG] Team %s[%d] already deleted", team.GetName(), team.GetID())
				continue
			}

			log.Printf("[DEBUG] Deleting team %s[%d]", team.GetSlug(), team.GetID())
			_, err := c.gh.Teams.DeleteTeamByID(ctx, c.org.GetID(), team.GetID())
			if err != nil {
				return err
			}
			deletedTeams[team.GetID()] = true

			log.Printf("[INFO] Deleted team %s", team.GetSlug())
		}

	}

	if c.org != nil {
		log.Printf("[INFO] Delete org %s", c.org.GetName())
		_, err := c.gh.Organizations.Delete(ctx, c.org.GetName())
		if err != nil {
			return err
		}
	}

	return nil

}

func setupGitHub(ctx context.Context, cfg *config) {

	c, err := NewClient(ctx, *cfg)
	if err != nil {
		log.Fatalf("[ERR] failed to create client: %v", err)
	}

	templateUsers := map[string]*TemplateUser{
		"user1": NewTemplateUser("user1"),
		"user2": NewTemplateUser("user2"),
		"user3": NewTemplateUser("user3"),
		"admin": NewTemplateUser("admin"),
	}
	users, err := c.OrgUsers(ctx)
	if err != nil {
		log.Fatalf("failed to load Org Users: %v\n", err)
	}
	for _, u := range users {
		name := u.GetLogin()

		if key, ok := userMap[name]; ok {
			log.Printf("[INFO] user match %q\n", name)
			templateUsers[key].User = u
		} else {
			log.Printf("[INFO] skip %q\n", name)
		}
	}

	// Create teams and assign membership
	teams := []struct {
		Name        string
		Description string
		MemberKeys  []string
	}{
		{"Public-All", "Team with All Members", []string{"user1", "user2", "user3", "admin"}},
		{"User1-Team", "Team with only user 1", []string{"user1"}},
		{"User2-Team", "Team with only user 2", []string{"user2"}},
		{"Mixed-Team", "Team with user 1, 2", []string{"user1", "user2"}},
	}
	var teamMap = make(map[string]*github.Team, 0)

	for _, t := range teams {
		team, err := c.GetOrCreateTeam(ctx, &github.NewTeam{
			Name:        t.Name,
			Description: &t.Description,
			Privacy:     strp("closed"),
		})
		if err != nil {
			log.Fatalf("[ERR] failed to get/create team %s: %v", t.Name, err)
		}

		for _, key := range t.MemberKeys {
			user := templateUsers[key]
			if err := c.AddTeamMembership(ctx, user.User, team); err != nil {
				log.Printf("[ERR] failed to add %q to team %v: %v", user.User.GetLogin(), team.GetName(), err)
				continue
			}
			templateUsers[key].Teams = append(templateUsers[key].Teams, team)
		}
		teamMap[team.GetName()] = team
	}

	// Create repos and assign to teams
	repos := []struct {
		Name     string
		TeamKeys []*github.Team
	}{
		{"repo-public", []*github.Team{teamMap["Public-All"]}},
		{"repo-private1", []*github.Team{teamMap["User1-Team"]}},
		{"repo-private2", []*github.Team{teamMap["User2-Team"]}},
		{"repo-private3", []*github.Team{teamMap["Mixed-Team"]}},
	}

	teamRepoMap := make(map[string]*github.Repository, 0)

	for _, r := range repos {
		for _, t := range r.TeamKeys {
			repo, err := c.GetOrCreateRepo(ctx, t, &github.Repository{
				Name:   strp(r.Name),
				TeamID: t.ID,
			})

			if err != nil {
				log.Printf("[WARN] failed to get/create repo %s", r.Name)
				continue
			}

			log.Printf("[DEBUG] team name: %s", *t.Name)
			teamRepoMap[*t.Name] = repo
		}
	}

	for _, v := range templateUsers {
		fmt.Println(v.String())
	}

	err = c.Cleanup(ctx, templateUsers, teamRepoMap)
	if err != nil {
		log.Printf("[ERR] cleanup failed: %v", err)
	}
}

// tokenAuthTransport adds token header authentication to requests.
type tokenAuthTransport struct {
	token   string
	wrapped http.RoundTripper
}

func (t *tokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf(`token-sudo user="%s",token="%s"`, "sourcegraph", t.token))
	return t.wrapped.RoundTrip(req)
}

type Q struct {
	CurrentUser struct {
		Username graphql.String
	}
}
type AddExternalServiceInput struct {
	Kind        graphql.String `json:"kind"`
	DisplayName graphql.String `json:"displayName"`
	Config      graphql.String `json:"config"`
}

type Mutation struct {
	input AddExternalServiceInput `graphql:"addExternalService(input: $input)"`
}

//	curl \
//	  -H 'Authorization: token-sudo user="SUDO-TO-USERNAME",token="sgp_8986cd7f808e6373ed7d499cafb23dd7ee4c481a"' \
//	  -d '{"query":"query { currentUser { username } }"}' \
func setupSourcegraph(ctx context.Context, cfg *config) error {
	client := graphql.NewClient(cfg.Sourcegraph.URL, &http.Client{
		Transport: &tokenAuthTransport{
			token:   cfg.Sourcegraph.Token,
			wrapped: http.DefaultTransport,
		},
	})
	q := Q{}
	err := client.Query(ctx, &q, nil)
	if err != nil {
		return err
	}
	log.Printf("RAWR: %v\n", q)

	extSvc := Mutation{}
	confJSON, err := json.Marshal(struct {
		URL   string   `json:"url"`
		Token string   `json:"token"`
		Org   []string `json:"orgs"`
	}{
		URL:   cfg.CodeHost.URL,
		Token: cfg.CodeHost.Token,
		Org:   []string{"william-templates"},
	})
	if err != nil {
		return err
	}
	input := AddExternalServiceInput{
		Kind:        "GITHUB",
		DisplayName: "TESTING",
		Config:      graphql.String(confJSON),
	}
	return client.Mutate(ctx, &extSvc, map[string]any{"input": input})
}

func main() {
	ctx := context.Background()
	cfg, err := Load("config.json")
	if err != nil {
		log.Fatalf("[ERR] failed to load config.json: %v\n", err)
	}
	// setupGitHub()
	if err := setupSourcegraph(ctx, cfg); err != nil {
		log.Fatalf("[ERR] failed to setup sourcegraph: %v\n", err)
	}

}
