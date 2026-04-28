package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"budget-helper/internal/domain"
	"github.com/rs/zerolog"
)

type SQLiteRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewSQLiteRepository(db *sql.DB, logger zerolog.Logger) *SQLiteRepository {
	return &SQLiteRepository{db: db, logger: logger}
}

func (r *SQLiteRepository) Add(ctx context.Context, tx domain.Transaction) error {
	var err error
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		err = r.execAdd(ctx, tx)
	}
	return err
}

func (r *SQLiteRepository) ListByFilter(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	var (
		list []domain.Transaction
		err  error
	)
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		list, err = r.execList(ctx, filter)
	}
	return list, err
}

func (r *SQLiteRepository) Upsert(ctx context.Context, budget domain.Budget) error {
	var err error
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		err = r.execUpsert(ctx, budget)
	}
	return err
}

func (r *SQLiteRepository) ListByMonth(ctx context.Context, monthStart time.Time) ([]domain.Budget, error) {
	var (
		list []domain.Budget
		err  error
	)
	if ctx == nil {
		err = errors.New("context is required")
	}
	if err == nil {
		list, err = r.execListBudgets(ctx, monthStart)
	}
	return list, err
}

func (r *SQLiteRepository) execAdd(ctx context.Context, tx domain.Transaction) error {
	var err error
	const q = `INSERT INTO transactions (
		id, amount_minor, currency, category, description, occurred_at, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.ExecContext(ctx, q, tx.ID, tx.AmountMinor, tx.Currency, tx.Category, tx.Description, tx.OccurredAt, tx.CreatedAt)
	r.logResult("insert_transaction", err)
	return err
}

func (r *SQLiteRepository) execList(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	var (
		rows *sql.Rows
		list []domain.Transaction
		err  error
	)
	query, args := buildListQuery(filter)
	r.logger.Debug().Str("op", "list_transactions").Str("query", query).Msg("db query")
	rows, err = r.db.QueryContext(ctx, query, args...)
	if err == nil {
		defer rows.Close()
		list, err = scanTransactions(rows)
	}
	if err != nil {
		r.logResult("list_transactions", err)
	}
	return list, err
}

func (r *SQLiteRepository) execUpsert(ctx context.Context, budget domain.Budget) error {
	var err error
	const q = `INSERT INTO budgets (
		id, category, currency, amount_minor, month_start, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(category, currency, month_start)
	DO UPDATE SET amount_minor = excluded.amount_minor, updated_at = excluded.updated_at`
	_, err = r.db.ExecContext(
		ctx,
		q,
		budget.ID,
		budget.Category,
		budget.Currency,
		budget.AmountMinor,
		budget.MonthStart,
		budget.CreatedAt,
		budget.UpdatedAt,
	)
	r.logResult("upsert_budget", err)
	return err
}

func (r *SQLiteRepository) execListBudgets(ctx context.Context, monthStart time.Time) ([]domain.Budget, error) {
	var (
		rows *sql.Rows
		list []domain.Budget
		err  error
	)
	const q = `SELECT id, category, currency, amount_minor, month_start, created_at, updated_at
		FROM budgets
		WHERE month_start = ?
		ORDER BY category`
	rows, err = r.db.QueryContext(ctx, q, monthStart)
	if err == nil {
		defer rows.Close()
		list, err = scanBudgets(rows)
	}
	if err != nil {
		r.logResult("list_budgets", err)
	}
	return list, err
}

func buildListQuery(filter domain.TransactionFilter) (string, []any) {
	base := `SELECT id, amount_minor, currency, category, description, occurred_at, created_at FROM transactions`
	conds := []string{"occurred_at >= ?", "occurred_at < ?"}
	args := []any{filter.MonthStart, filter.MonthStart.AddDate(0, 1, 0)}
	if filter.Category != "" {
		conds = append(conds, "category = ?")
		args = append(args, filter.Category)
	}
	if filter.Search != "" {
		conds = append(conds, "description LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}
	query := fmt.Sprintf("%s WHERE %s ORDER BY occurred_at DESC", base, strings.Join(conds, " AND "))
	return query, args
}

func scanTransactions(rows *sql.Rows) ([]domain.Transaction, error) {
	var (
		list []domain.Transaction
		err  error
	)
	for rows.Next() {
		var tx domain.Transaction
		err = rows.Scan(&tx.ID, &tx.AmountMinor, &tx.Currency, &tx.Category, &tx.Description, &tx.OccurredAt, &tx.CreatedAt)
		if err != nil {
			break
		}
		list = append(list, tx)
	}
	if err == nil {
		err = rows.Err()
	}
	return list, err
}

func scanBudgets(rows *sql.Rows) ([]domain.Budget, error) {
	var (
		list []domain.Budget
		err  error
	)
	for rows.Next() {
		var b domain.Budget
		err = rows.Scan(&b.ID, &b.Category, &b.Currency, &b.AmountMinor, &b.MonthStart, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			break
		}
		list = append(list, b)
	}
	if err == nil {
		err = rows.Err()
	}
	return list, err
}

func (r *SQLiteRepository) logResult(op string, err error) {
	if err != nil {
		r.logger.Error().Str("op", op).Err(err).Msg("db error")
	}
	if err == nil {
		r.logger.Debug().Str("op", op).Msg("db ok")
	}
}
