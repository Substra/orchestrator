package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// TaskRegisteredTotal keeps track of registered tasks
	TaskRegisteredTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orc_task_registered_total",
			Help: "Number of registered tasks",
		},
		[]string{"channel", "category"},
	)

	// TaskRegistrationBatchSize monitors the usual batch size when registering tasks
	TaskRegistrationBatchSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "orc_task_registration_batch_size",
			Help:    "Size of task batches",
			Buckets: []float64{5, 10, 50, 100, 500, 1000, 5000, 10000},
		},
		[]string{"channel"},
	)

	// TaskUpdatedTotal counts the number of task updates
	TaskUpdatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orc_task_updated_total",
			Help: "Number of updated tasks",
		},
		[]string{"channel", "status"},
	)

	// TaskUpdateCascadeSize keeps track of how many tasks are updated in cascade
	TaskUpdateCascadeSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "orc_task_update_cascade_size",
			Help:    "Number of updated children as a consequence of a task update",
			Buckets: []float64{5, 10, 50, 100, 500, 1000, 5000, 10000},
		},
		[]string{"channel", "status"},
	)
)

func init() {
	prometheus.MustRegister(TaskRegisteredTotal)
	prometheus.MustRegister(TaskRegistrationBatchSize)
	prometheus.MustRegister(TaskUpdatedTotal)
	prometheus.MustRegister(TaskUpdateCascadeSize)
}
