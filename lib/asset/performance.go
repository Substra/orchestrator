package asset

import "strings"

// GetKey returns the performance key derived from its referenced task and metric.
func (p *Performance) GetKey() string {
	return strings.Join([]string{p.ComputeTaskKey, p.MetricKey}, "|")
}
