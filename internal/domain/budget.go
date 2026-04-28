package domain

import "time"

type Budget struct {
	ID          string
	Category    string
	Currency    string
	AmountMinor int64
	MonthStart  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
