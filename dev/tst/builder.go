package tst

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/google/uuid"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ScenarioResource struct {
	name string
	id   string
	key  string
}

func NewScenarioResource(name string) *ScenarioResource {
	id := id()
	key := joinID(name, "-", id, 39)
	return &ScenarioResource{
		name: name,
		id:   id,
		key:  key,
	}

}

func (s *ScenarioResource) ID() string {
	return s.id
}

func (s *ScenarioResource) Name() string {
	return s.name
}

func (s *ScenarioResource) Key() string {
	return s.key
}

type GitHubScenarioBuilder struct {
	client          *GitHubClient
	store           *scenarioStore
	setupActions    []Action
	teardownActions []Action
}

type Scenario interface {
}

type scenario struct {
	client *GitHubClient
	users  []*github.User
	teams  []*github.Team
	repos  []*github.Repository
	org    *github.Organization
}

type ActionResult interface {
	Get() any
}

type ActionFn func(ctx context.Context, store *scenarioStore) (ActionResult, error)

type Action interface {
	Name() string
	Hash() []byte
	Complete() bool
	Do(ctx context.Context, store *scenarioStore) (ActionResult, error)
}

type action struct {
	name     string
	hash     []byte
	complete bool
	doFn     ActionFn
}

func (a *action) Do(ctx context.Context, store *scenarioStore) (ActionResult, error) {
	result, err := a.doFn(ctx, store)
	a.complete = true
	return result, err
}

func (a *action) Hash() []byte {
	return a.hash
}

func (a *action) Name() string {
	return a.name
}

func (a *action) Complete() bool {
	return a.complete
}

type GitHubClient struct {
	cfg *CodeHost
	c   *github.Client
}

type actionResult[T any] struct {
	item T
}

func (a *actionResult[T]) Get() any {
	return a.item
}

var _scenario Scenario = &scenario{}

func NewGitHubScenario(ctx context.Context, cfg Config) (*GitHubScenarioBuilder, error) {
	client, err := NewGitHubClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &GitHubScenarioBuilder{
		client:          client,
		store:           NewStore(),
		setupActions:    make([]Action, 0),
		teardownActions: make([]Action, 0),
	}, nil
}

func (sb *GitHubScenarioBuilder) Org(name string) *GitHubScenarioBuilder {
	org := NewGitHubScenarioOrg(name)
	sb.setupActions = append(sb.setupActions, org.CreateOrgAction(sb.client))
	sb.teardownActions = append(sb.teardownActions, org.DeleteOrgAction(sb.client))
	return sb
}

func (sb *GitHubScenarioBuilder) Users(users ...GitHubScenarioUser) *GitHubScenarioBuilder {
	for _, u := range users {
		if u == Admin {
			sb.setupActions = append(sb.setupActions, u.GetUserAction(sb.client))
		} else {
			sb.setupActions = append(
				sb.setupActions,
				u.CreateUserAction(sb.client))
			sb.teardownActions = append(
				sb.teardownActions,
				u.DeleteUserAction(sb.client))
		}
	}
	return sb
}

func Team(name string, u ...GitHubScenarioUser) *GitHubScenarioTeam {
	return NewGitHubScenarioTeam(name, u...)
}

func (sb *GitHubScenarioBuilder) Teams(teams ...*GitHubScenarioTeam) *GitHubScenarioBuilder {
	for _, t := range teams {
		sb.setupActions = append(sb.setupActions, t.CreateTeamAction(sb.client))
		sb.setupActions = append(sb.setupActions, t.AssignTeamAction(sb.client))
		sb.teardownActions = append(sb.teardownActions, t.DeleteTeamAction(sb.client))
	}

	return sb
}

func (sb *GitHubScenarioBuilder) Repos(repos ...*GitHubScenarioRepo) *GitHubScenarioBuilder {
	for _, r := range repos {
		if r.fork {
			sb.setupActions = append(sb.setupActions, r.ForkRepoAction(sb.client))
			sb.setupActions = append(sb.setupActions, r.GetForkedRepo(sb.client))
			// Seems like you can't change permissions for a repo fork
			//sb.setupActions = append(sb.setupActions, r.SetPermissionsAction(sb.client))
			sb.teardownActions = append(sb.teardownActions, r.DeleteRepoAction(sb.client))
		}
		sb.setupActions = append(sb.setupActions, r.AssignTeamAction(sb.client))
	}

	return sb
}

func PublicRepo(name string, team string, fork bool) *GitHubScenarioRepo {
	return NewGitHubScenarioRepo(name, team, fork, false)
}

func PrivateRepo(name string, team string, fork bool) *GitHubScenarioRepo {
	return NewGitHubScenarioRepo(name, team, fork, true)
}

func (sb *GitHubScenarioBuilder) setupPlan() string {
	b := strings.Builder{}
	for _, act := range sb.setupActions {
		b.WriteString(act.Name())
		b.WriteByte('\n')
	}

	return b.String()
}

func (sb *GitHubScenarioBuilder) tearDownPlan() string {
	b := strings.Builder{}
	actions := sb.teardownActions
	for i := len(actions) - 1; i >= 0; i-- {
		b.WriteString(actions[i].Name())
		b.WriteByte('\n')
	}

	return b.String()
}

func (sb *GitHubScenarioBuilder) String() string {
	b := strings.Builder{}
	b.WriteString("Setup\n")
	b.WriteString("======\n")
	b.WriteString(sb.setupPlan())
	b.WriteByte('\n')
	b.WriteString("Teardown\n")
	b.WriteString("========\n")
	b.WriteString(sb.tearDownPlan())
	return b.String()
}

func id() string {
	id := []byte(uuid.NewString())
	return base64.RawStdEncoding.EncodeToString(id[:])

}

func joinID(v, sep, id string, max int) string {
	length := int(math.Min(float64(len(id)), float64(max-len(sep)-len(v))))
	return v + sep + id[:length]
}

func applyActions(ctx context.Context, store *scenarioStore, actions []Action, failFast bool) error {
	setupStart := time.Now().UTC()
	var errs errors.MultiError
	for _, action := range actions {
		fmt.Printf("Applying '%s' = ", action.Name())
		now := time.Now().UTC()

		var err error
		if !action.Complete() {
			_, err = action.Do(ctx, store)
		} else {
			fmt.Print("[SKIPPED]\n")
			continue
		}

		duration := time.Now().UTC().Sub(now)
		if err != nil {
			if failFast {
				fmt.Printf("[FAILED] (%s)\n", duration.String())
				return err
			} else {
				fmt.Printf("[FAILED] (%s)\n", duration.String())
				errs = errors.Append(errs, err)
			}
		} else {
			fmt.Printf("[SUCCESS] (%s)\n", duration.String())
		}
	}
	fmt.Printf("Run complete: %s\n", time.Now().UTC().Sub(setupStart))
	return errs
}

func (sb *GitHubScenarioBuilder) Setup(ctx context.Context) (Scenario, func(context.Context) error, error) {
	fmt.Println("-- Setup --")
	err := applyActions(ctx, sb.store, sb.setupActions, true)
	return scenario{}, sb.TearDown, err
}

func (sb *GitHubScenarioBuilder) TearDown(ctx context.Context) error {
	reversed := make([]Action, 0, len(sb.teardownActions))
	for i := len(sb.teardownActions) - 1; i >= 0; i-- {
		reversed = append(reversed, sb.teardownActions[i])
	}

	fmt.Println("-- Teardown --")
	return applyActions(ctx, sb.store, reversed, false)
}

func strp(v string) *string {
	return &v
}
