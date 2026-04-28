package domain

import "time"

type TransactionFilter struct {
	MonthStart time.Time
	Category   string
	Search     string
}
