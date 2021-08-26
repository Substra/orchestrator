package service

import "time"

type TimeAPI interface {
	GetTransactionTime() time.Time
}

type TimeServiceProvider interface {
	GetTimeService() TimeAPI
}

type TimeService struct {
	t time.Time
}

func NewTimeService(t time.Time) *TimeService {
	return &TimeService{t}
}

func (ts *TimeService) GetTransactionTime() time.Time {
	return ts.t
}
