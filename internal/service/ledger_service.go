package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"budget-helper/internal/domain"
	"budget-helper/internal/report"
	"budget-helper/internal/repository"
)

type LedgerService struct {
	txRepo     repository.TransactionRepository
	budgetRepo repository.BudgetRepository
	now        func() time.Time
	idGen      func() (string, error)
}

type BudgetStatus struct {
	Category    string
	Currency    string
	BudgetMinor int64
	SpentMinor  int64
	Exceeded    bool
}

type Dashboard struct {
	MonthStart    time.Time
	Currency      string
	TotalIncome   int64
	TotalExpenses int64
	Net           int64
	TopCategories []domain.CategoryTotal
	ExpenseBars   []report.ExpenseBar
	Budgets       []BudgetStatus
}

func NewLedgerService(txRepo repository.TransactionRepository, budgetRepo repository.BudgetRepository, now func() time.Time, idGen func() (string, error)) *LedgerService {
	return &LedgerService{txRepo: txRepo, budgetRepo: budgetRepo, now: now, idGen: idGen}
}

func (s *LedgerService) AddTransaction(ctx context.Context, amountMinor int64, currency, category, description string, occurredAt time.Time) (domain.Transaction, error) {
	var (
		tx  domain.Transaction
		err error
	)
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		err = validateTransaction(amountMinor, currency, category)
	}
	if err == nil {
		var id string
		id, err = s.idGen()
		if err == nil {
			now := s.now()
			if occurredAt.IsZero() {
				occurredAt = now
			}
			tx = domain.Transaction{
				ID:          id,
				AmountMinor: amountMinor,
				Currency:    currency,
				Category:    category,
				Description: description,
				OccurredAt:  occurredAt,
				CreatedAt:   now,
			}
			err = s.txRepo.Add(ctx, tx)
		}
	}
	return tx, err
}

func (s *LedgerService) ListTransactions(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	var (
		list []domain.Transaction
		err  error
	)
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		err = validateFilter(filter)
	}
	if err == nil {
		list, err = s.txRepo.ListByFilter(ctx, filter)
	}
	return list, err
}

func (s *LedgerService) SetMonthlyBudget(ctx context.Context, year int, month time.Month, category, currency string, amountMinor int64) (domain.Budget, error) {
	var (
		budget domain.Budget
		err    error
	)
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		err = validateBudget(year, month, category, currency, amountMinor)
	}
	if err == nil {
		var id string
		id, err = s.idGen()
		if err == nil {
			now := s.now()
			budget = domain.Budget{
				ID:          id,
				Category:    category,
				Currency:    currency,
				AmountMinor: amountMinor,
				MonthStart:  monthStart(year, month),
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			err = s.budgetRepo.Upsert(ctx, budget)
		}
	}
	return budget, err
}

func (s *LedgerService) Dashboard(ctx context.Context, year int, month time.Month) (Dashboard, error) {
	var (
		data Dashboard
		err  error
	)
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		var list []domain.Transaction
		filter := domain.TransactionFilter{MonthStart: monthStart(year, month)}
		list, err = s.ListTransactions(ctx, filter)
		if err == nil {
			data = buildDashboard(list)
			data.MonthStart = filter.MonthStart
			data.Budgets, err = s.buildBudgetStatus(ctx, filter.MonthStart, list)
		}
	}
	return data, err
}

func (s *LedgerService) MonthlyReport(ctx context.Context, year int, month time.Month) (report.MonthlyReport, error) {
	var (
		out  report.MonthlyReport
		err  error
		list []domain.Transaction
	)
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		filter := domain.TransactionFilter{MonthStart: monthStart(year, month)}
		list, err = s.ListTransactions(ctx, filter)
	}
	if err == nil {
		out = report.BuildMonthlyReport(list)
	}
	return out, err
}

func (s *LedgerService) ExportMonthlyReport(ctx context.Context, year int, month time.Month, w io.Writer) error {
	var err error
	var rep report.MonthlyReport
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		rep, err = s.MonthlyReport(ctx, year, month)
	}
	if err == nil {
		err = report.ExportCSV(w, rep)
	}
	return err
}

func (s *LedgerService) buildBudgetStatus(ctx context.Context, monthStart time.Time, list []domain.Transaction) ([]BudgetStatus, error) {
	var (
		budgets []domain.Budget
		err     error
		status  []BudgetStatus
	)
	budgets, err = s.budgetRepo.ListByMonth(ctx, monthStart)
	if err == nil {
		status = budgetStatus(budgets, list)
	}
	return status, err
}

func validateTransaction(amountMinor int64, currency, category string) error {
	var err error
	if amountMinor == 0 {
		err = errors.New("amount must not be zero")
	}
	if err == nil && strings.TrimSpace(currency) == "" {
		err = errors.New("currency is required")
	}
	if err == nil && strings.TrimSpace(category) == "" {
		err = errors.New("category is required")
	}
	return err
}

func validateFilter(filter domain.TransactionFilter) error {
	var err error
	if filter.MonthStart.IsZero() {
		err = errors.New("month is required")
	}
	return err
}

func validateBudget(year int, month time.Month, category, currency string, amountMinor int64) error {
	var err error
	if year < 1 || month < time.January || month > time.December {
		err = errors.New("invalid month")
	}
	if err == nil && strings.TrimSpace(category) == "" {
		err = errors.New("category is required")
	}
	if err == nil && strings.TrimSpace(currency) == "" {
		err = errors.New("currency is required")
	}
	if err == nil && amountMinor <= 0 {
		err = errors.New("budget must be positive")
	}
	return err
}

func monthStart(year int, month time.Month) time.Time {
	loc := time.Local
	return time.Date(year, month, 1, 0, 0, 0, 0, loc)
}

func buildDashboard(list []domain.Transaction) Dashboard {
	data := Dashboard{}
	rep := report.BuildMonthlyReport(list)
	data.Currency = rep.Currency
	data.TotalIncome = rep.TotalIncome
	data.TotalExpenses = rep.TotalExpenses
	data.Net = rep.Net
	data.TopCategories = rep.TopCategories
	data.ExpenseBars = report.BuildExpenseBars(rep.TopCategories, 24)
	return data
}

func budgetStatus(budgets []domain.Budget, list []domain.Transaction) []BudgetStatus {
	status := make([]BudgetStatus, 0, len(budgets))
	expenses := expenseTotals(list)
	for _, b := range budgets {
		spent := expenses[b.Category]
		exceeded := exceedsBudget(spent, b.AmountMinor)
		status = append(status, BudgetStatus{Category: b.Category, Currency: b.Currency, BudgetMinor: b.AmountMinor, SpentMinor: spent, Exceeded: exceeded})
	}
	return status
}

func expenseTotals(list []domain.Transaction) map[string]int64 {
	out := map[string]int64{}
	for _, tx := range list {
		if tx.AmountMinor < 0 {
			out[tx.Category] += tx.AmountMinor
		}
	}
	return out
}

func exceedsBudget(spentMinor, budgetMinor int64) bool {
	spent := spentMinor
	if spent < 0 {
		spent = -spent
	}
	return spent > budgetMinor
}
