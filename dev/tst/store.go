package tst

import (
	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var castFailure error

type scenarioStore struct {
	store map[string]any
}

func NewStore() *scenarioStore {
	return &scenarioStore{
		store: make(map[string]any),
	}
}

func (s *scenarioStore) SetOrg(org *github.Organization) {
	s.store["org"] = org
}
func (s *scenarioStore) GetOrg() (*github.Organization, error) {
	var result *github.Organization
	if v, ok := s.store["org"]; ok {
		if t, ok := v.(*github.Organization); ok {
			result = t
		} else {
			return result, castFailure
		}
	} else {
		return result, errors.Newf("%s not found - it might not have been loaded yet", "org")
	}
	return result, nil
}

func (s *scenarioStore) SetScenarioUserMapping(u *GitHubScenarioUser, user *github.User) {
	s.store[u.Key()] = user
}

func (s *scenarioStore) SetUsers(users []*github.User) {
	s.store["all-users"] = users
}

func (s *scenarioStore) GetUsers() ([]*github.User, error) {
	var result []*github.User
	if v, ok := s.store["org"]; ok {
		if t, ok := v.([]*github.User); ok {
			result = t
		} else {
			return result, castFailure
		}
	} else {
		return result, errors.Newf("%s not found - it might not have been loaded yet", "all-users")
	}
	return result, nil
}

func (s *scenarioStore) GetScenarioUser(u GitHubScenarioUser) (*github.User, error) {
	var result *github.User
	if v, ok := s.store[u.Key()]; ok {
		if t, ok := v.(*github.User); ok {
			result = t
		} else {
			return result, castFailure
		}
	} else {
		return result, errors.Newf("%s not found - it might not have been loaded yet", u.Key())
	}
	return result, nil
}

func (s *scenarioStore) SetTeam(gt *GitHubScenarioTeam, t *github.Team) {
	// Store it twice so we have two ways of retrieving a team
	s.store[gt.Name()] = t
	s.store[gt.Key()] = t

}

func (s *scenarioStore) GetTeamByName(name string) (*github.Team, error) {
	var result *github.Team
	if v, ok := s.store[name]; ok {
		if t, ok := v.(*github.Team); ok {
			result = t
		} else {
			return result, castFailure
		}
	} else {
		return result, errors.Newf("%s not found - it might not have been loaded yet", name)
	}
	return result, nil
}

func (s *scenarioStore) GetTeam(t *GitHubScenarioTeam) (*github.Team, error) {
	return s.GetTeamByName(t.Name())
}

func (s *scenarioStore) SetRepo(r *GitHubScenarioRepo, repo *github.Repository) {
	s.store[r.Key()] = repo
}

func (s *scenarioStore) GetRepo(r *GitHubScenarioRepo) (*github.Repository, error) {
	var result *github.Repository
	if v, ok := s.store[r.Key()]; ok {
		if t, ok := v.(*github.Repository); ok {
			result = t
		} else {
			return result, castFailure
		}
	} else {
		return result, errors.Newf("%s not found - it might not have been loaded yet", r.Key())
	}
	return result, nil
}
