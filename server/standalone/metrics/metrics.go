package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// DBTransactionTotal keeps track of the number of transactions
	DBTransactionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orc_db_transaction_total",
			Help: "Number of database transactions, partitioned by method and outcome (commit/rollback)",
		},
		[]string{"method", "outcome"},
	)

	// EventDispatchedTotal keeps track of the number of dispatched events
	EventDispatchedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orc_event_sent_total",
			Help: "Number of events dispatched",
		},
	)
)

func init() {
	prometheus.MustRegister(DBTransactionTotal)
	prometheus.MustRegister(EventDispatchedTotal)
}
