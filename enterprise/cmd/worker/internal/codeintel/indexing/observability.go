package indexing

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type dependencyReposOperations struct {
	HandleDependencySyncing       *observation.Operation
	HandleDependencyIndexing      *observation.Operation
	InsertCloneableDependencyRepo *observation.Operation
}

var (
	once               sync.Once
	dependencyReposOps *dependencyReposOperations
)

func newOperations(observationContext *observation.Context) *dependencyReposOperations {
	once.Do(func() {
		m := metrics.NewREDMetrics(
			observationContext.Registerer,
			"codeintel_dependency_repos",
			metrics.WithLabels("op", "scheme", "new"),
		)

		opWithMetrics := func(prefix, name string) *observation.Operation {
			return observationContext.Operation(observation.Op{
				Name:              fmt.Sprintf("codeintel.%s.%s", prefix, name),
				MetricLabelValues: []string{name},
				Metrics:           m,
			})
		}
		opSansMetrics := func(prefix, name string) *observation.Operation {
			return observationContext.Operation(observation.Op{
				Name: fmt.Sprintf("codeintel.%s.%s", prefix, name),
			})
		}

		dependencyReposOps = &dependencyReposOperations{
			HandleDependencySyncing:       opSansMetrics("dependencyrepos", "HandleDependencySyncing"),
			HandleDependencyIndexing:      opSansMetrics("dependencyrepos", "HandleDependencyIndexing"),
			InsertCloneableDependencyRepo: opWithMetrics("dependencyrepos", "InsertCloneableDependencyRepo"),
		}
	})
	return dependencyReposOps
}
