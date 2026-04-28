package service

import (
	"context"
	"testing"
	"time"

	"budget-helper/internal/domain"
)

type fakeTxRepo struct {
	added []domain.Transaction
	list  []domain.Transaction
	err   error
}

func (f *fakeTxRepo) Add(ctx context.Context, tx domain.Transaction) error {
	f.added = append(f.added, tx)
	return f.err
}

func (f *fakeTxRepo) ListByFilter(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	return f.list, f.err
}

type fakeBudgetRepo struct {
	budgets []domain.Budget
	err     error
}

func (f *fakeBudgetRepo) Upsert(ctx context.Context, budget domain.Budget) error {
	f.budgets = append(f.budgets, budget)
	return f.err
}

func (f *fakeBudgetRepo) ListByMonth(ctx context.Context, monthStart time.Time) ([]domain.Budget, error) {
	return f.budgets, f.err
}

type addCase struct {
	name     string
	amount   int64
	currency string
	category string
	wantErr  bool
}

func TestAddTransaction(t *testing.T) {
	cases := []addCase{
		{name: "valid", amount: 1000, currency: "USD", category: "Salary", wantErr: false},
		{name: "zero", amount: 0, currency: "USD", category: "Misc", wantErr: true},
		{name: "no currency", amount: 1000, currency: "", category: "Misc", wantErr: true},
		{name: "no category", amount: 1000, currency: "USD", category: "", wantErr: true},
	}
	for _, tc := range cases {
		runAddCase(t, tc)
	}
}

func runAddCase(t *testing.T, tc addCase) {
	now := time.Date(2026, time.February, 1, 10, 0, 0, 0, time.UTC)
	txRepo := &fakeTxRepo{}
	budgetRepo := &fakeBudgetRepo{}
	idGen := func() (string, error) { return "id-1", nil }
	svc := NewLedgerService(txRepo, budgetRepo, func() time.Time { return now }, idGen)
	_, err := svc.AddTransaction(context.Background(), tc.amount, tc.currency, tc.category, "", time.Time{})
	if tc.wantErr && err == nil {
		t.Fatalf("expected error")
	}
	if !tc.wantErr && err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type budgetCase struct {
	name      string
	budgets   []domain.Budget
	list      []domain.Transaction
	category  string
	exceeded  bool
	spent     int64
	budgetAmt int64
}

func TestBudgetStatus(t *testing.T) {
	cases := []budgetCase{
		{
			name: "exceeded",
			budgets: []domain.Budget{{Category: "Food", Currency: "USD", AmountMinor: 1000}},
			list: []domain.Transaction{{Category: "Food", AmountMinor: -1500}},
			category: "Food",
			exceeded: true,
			spent: -1500,
			budgetAmt: 1000,
		},
		{
			name: "within",
			budgets: []domain.Budget{{Category: "Food", Currency: "USD", AmountMinor: 2000}},
			list: []domain.Transaction{{Category: "Food", AmountMinor: -1500}},
			category: "Food",
			exceeded: false,
			spent: -1500,
			budgetAmt: 2000,
		},
	}
	for _, tc := range cases {
		runBudgetCase(t, tc)
	}
}

func runBudgetCase(t *testing.T, tc budgetCase) {
	status := budgetStatus(tc.budgets, tc.list)
	if len(status) != 1 {
		t.Fatalf("expected 1 status")
	}
	if status[0].Category != tc.category || status[0].Exceeded != tc.exceeded {
		t.Fatalf("status mismatch")
	}
	if status[0].SpentMinor != tc.spent || status[0].BudgetMinor != tc.budgetAmt {
		t.Fatalf("amount mismatch")
	}
}

func TestMonthlyReport(t *testing.T) {
	txRepo := &fakeTxRepo{list: sampleTransactions()}
	budgetRepo := &fakeBudgetRepo{}
	idGen := func() (string, error) { return "id-1", nil }
	svc := NewLedgerService(txRepo, budgetRepo, time.Now, idGen)
	report, err := svc.MonthlyReport(context.Background(), 2026, time.February)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalIncome != 120000 || report.TotalExpenses != -45000 || report.Net != 75000 {
		t.Fatalf("totals mismatch")
	}
}

func sampleTransactions() []domain.Transaction {
	now := time.Date(2026, time.February, 1, 10, 0, 0, 0, time.UTC)
	return []domain.Transaction{
		{ID: "1", AmountMinor: 100000, Currency: "USD", Category: "Salary", OccurredAt: now, CreatedAt: now},
		{ID: "2", AmountMinor: -30000, Currency: "USD", Category: "Food", OccurredAt: now, CreatedAt: now},
		{ID: "3", AmountMinor: -15000, Currency: "USD", Category: "Transport", OccurredAt: now, CreatedAt: now},
		{ID: "4", AmountMinor: -5000, Currency: "USD", Category: "Food", OccurredAt: now, CreatedAt: now},
		{ID: "5", AmountMinor: 20000, Currency: "USD", Category: "Gift", OccurredAt: now, CreatedAt: now},
	}
}
