package repos

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestParseGitHubExcludeRule(t *testing.T) {
	rule := &schema.ExcludedGitHubRepo{Stars: "< 100", Size: ">= 1GB"}

	fn, err := ParseGitHubExcludeRule(rule)
	assert.Nil(t, err)

	assert.True(t, fn(github.Repository{StargazerCount: 99, SizeKibiBytes: 976562 + 1}))
}
