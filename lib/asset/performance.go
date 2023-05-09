package asset

import "strings"

// GetKey returns the performance key derived from its referenced task, metric and identifier.
func (p *Performance) GetKey() string {
	return strings.Join([]string{p.ComputeTaskKey, p.MetricKey, p.ComputeTaskOutputIdentifier}, "|")
}
