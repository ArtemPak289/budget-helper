package domain

import "time"

type Transaction struct {
	ID          string
	AmountMinor int64
	Currency    string
	Category    string
	Description string
	OccurredAt  time.Time
	CreatedAt   time.Time
}
