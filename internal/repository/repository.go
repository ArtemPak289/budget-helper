package repository

import (
	"context"
	"time"

	"budget-helper/internal/domain"
)

type TransactionRepository interface {
	Add(ctx context.Context, tx domain.Transaction) error
	ListByFilter(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error)
}

type BudgetRepository interface {
	Upsert(ctx context.Context, budget domain.Budget) error
	ListByMonth(ctx context.Context, monthStart time.Time) ([]domain.Budget, error)
}
