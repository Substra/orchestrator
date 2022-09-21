package asset

func (c *ComputePlan) IsTerminated() bool {
	return c.CancelationDate != nil || c.FailureDate != nil
}
