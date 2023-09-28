package httpapi

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewCodeCompletionsHandler is an http handler which sends back code completion results.
func NewCodeCompletionsHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("code", "code completions handler")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureCode)
	return newCompletionsHandler(
		logger,
		types.CompletionsFeatureCode,
		rl,
		"code",
		func(requestParams types.CodyCompletionRequestParameters, c *conftypes.CompletionsConfig) (string, error) {
			// logg.Printf("### getModel: %q", c.CompletionModel)
			// logg.Printf("### requestParams.Model: %q", requestParams.Model)
			if isAllowedCustomModel(requestParams.Model) {
				// logg.Printf("#### NewCodeCompletionsHandler 1")
				return requestParams.Model, nil
			}
			if requestParams.Model != "" {
				// logg.Printf("#### NewCodeCompletionsHandler 2")
				return "", errors.New("Unsupported custom model")
			}
			// logg.Printf("#### NewCodeCompletionsHandler 3")
			return c.CompletionModel, nil
		},
	)
}

// We only allow dotcom clients to select a custom code model and maintain an allowlist for which
// custom values we support
func isAllowedCustomModel(model string) bool {
	// if !(envvar.SourcegraphDotComMode()) {
	// 	return false
	// }

	switch model {
	case "fireworks/accounts/fireworks/models/starcoder-16b-w8a16":
		fallthrough
	case "fireworks/accounts/fireworks/models/starcoder-7b-w8a16":
		fallthrough
	case "fireworks/accounts/fireworks/models/starcoder-3b-w8a16":
		fallthrough
	case "fireworks/accounts/fireworks/models/starcoder-1b-w8a16":
		fallthrough
	case "fireworks/accounts/fireworks/models/llama-v2-7b-code":
		fallthrough
	case "fireworks/accounts/fireworks/models/llama-v2-13b-code":
		fallthrough
	case "fireworks/accounts/fireworks/models/llama-v2-13b-code-instruct":
		fallthrough
	case "fireworks/accounts/fireworks/models/wizardcoder-15b":
		return true
	}

	return false
}
