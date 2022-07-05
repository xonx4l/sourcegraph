package indexing

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type dependencySyncingOperations struct {
	HandleDependencySyncing       *observation.Operation
	InsertCloneableDependencyRepo *observation.Operation
}

type dependencyIndexingOperations struct {
	HandleDependencyIndexing *observation.Operation
}

var (
	syncingOnce           sync.Once
	indexingOnce          sync.Once
	dependencySyncingOps  *dependencySyncingOperations
	dependencyIndexingOps *dependencyIndexingOperations
)

func newSyncingOperations(observationContext *observation.Context) *dependencySyncingOperations {
	syncingOnce.Do(func() {
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

		dependencySyncingOps = &dependencySyncingOperations{
			HandleDependencySyncing:       opSansMetrics("dependencyrepos", "HandleDependencySyncing"),
			InsertCloneableDependencyRepo: opWithMetrics("dependencyrepos", "InsertCloneableDependencyRepo"),
		}
	})
	return dependencySyncingOps
}

func newIndexingOperations(observationContext *observation.Context) *dependencyIndexingOperations {
	indexingOnce.Do(func() {
		opSansMetrics := func(prefix, name string) *observation.Operation {
			return observationContext.Operation(observation.Op{
				Name: fmt.Sprintf("codeintel.%s.%s", prefix, name),
			})
		}

		dependencyIndexingOps = &dependencyIndexingOperations{
			HandleDependencyIndexing: opSansMetrics("dependencyrepos", "HandleDependencyIndexing"),
		}
	})
	return dependencyIndexingOps
}
