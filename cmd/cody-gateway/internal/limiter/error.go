package limiter

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

type RateLimitExceededError struct {
	Limit      int64
	Feature    codygateway.Feature
	RetryAfter time.Time
}

// Error generates a simple string that is fairly static for use in logging.
// This helps with categorizing errors. For more detailed output use Summary().
func (e RateLimitExceededError) Error() string { return "rate limit exceeded" }

func (e RateLimitExceededError) Summary() string {
	return fmt.Sprintf(
		"your organization has used all %d allocated %s requests until %s, please contact your admin for increased allocation",
		e.Limit,
		e.Feature.DisplayName(),
		e.RetryAfter.Truncate(time.Second),
	)
}

func (e RateLimitExceededError) WriteResponse(w http.ResponseWriter) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.FormatInt(e.Limit, 10))
	w.Header().Set("x-ratelimit-remaining", "0")
	w.Header().Set("retry-after", e.RetryAfter.UTC().Format(time.RFC1123))
	// Use Summary instead of Error for more informative text
	http.Error(w, e.Summary(), http.StatusTooManyRequests)
}

type NoAccessError struct {
	feature codygateway.Feature
}

func (e NoAccessError) Error() string {
	return fmt.Sprintf("%s access has not been granted", e.feature.DisplayName())
}
