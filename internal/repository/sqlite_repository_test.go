package repository

import (
	"context"
	"io"
	"testing"
	"time"

	"budget-helper/internal/domain"
	"github.com/rs/zerolog"
)

func TestSQLiteRepositoryIntegration(t *testing.T) {
	dsn := "file:memdb1?mode=memory&cache=shared"
	db, err := OpenSQLite(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	err = Migrate(db)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	log := zerolog.New(io.Discard)
	repo := NewSQLiteRepository(db, log)
	monthStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	tx := domain.Transaction{
		ID:          "id-1",
		AmountMinor: -5000,
		Currency:    "USD",
		Category:    "Food",
		Description: "Lunch",
		OccurredAt:  monthStart.Add(2 * time.Hour),
		CreatedAt:   monthStart,
	}
	err = repo.Add(context.Background(), tx)
	if err != nil {
		t.Fatalf("add tx: %v", err)
	}
	filter := domain.TransactionFilter{MonthStart: monthStart}
	list, err := repo.ListByFilter(context.Background(), filter)
	if err != nil {
		t.Fatalf("list tx: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 transaction")
	}
	budget := domain.Budget{
		ID:          "b-1",
		Category:    "Food",
		Currency:    "USD",
		AmountMinor: 10000,
		MonthStart:  monthStart,
		CreatedAt:   monthStart,
		UpdatedAt:   monthStart,
	}
	err = repo.Upsert(context.Background(), budget)
	if err != nil {
		t.Fatalf("upsert budget: %v", err)
	}
	budgets, err := repo.ListByMonth(context.Background(), monthStart)
	if err != nil {
		t.Fatalf("list budgets: %v", err)
	}
	if len(budgets) != 1 {
		t.Fatalf("expected 1 budget")
	}
}
