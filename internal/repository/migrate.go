package repository

import (
	"database/sql"
	"errors"
)

func Migrate(db *sql.DB) error {
	var err error
	if db == nil {
		err = errors.New("db is required")
	}
	if err == nil {
		err = execSchema(db)
	}
	return err
}

func execSchema(db *sql.DB) error {
	var err error
	const txTable = `CREATE TABLE IF NOT EXISTS transactions (
		id TEXT PRIMARY KEY,
		amount_minor INTEGER NOT NULL,
		currency TEXT NOT NULL,
		category TEXT NOT NULL,
		description TEXT,
		occurred_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL
	)`
	_, err = db.Exec(txTable)
	if err == nil {
		const budgetTable = `CREATE TABLE IF NOT EXISTS budgets (
			id TEXT PRIMARY KEY,
			category TEXT NOT NULL,
			currency TEXT NOT NULL,
			amount_minor INTEGER NOT NULL,
			month_start TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			UNIQUE(category, currency, month_start)
		)`
		_, err = db.Exec(budgetTable)
	}
	if err == nil {
		const idx = `CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at ON transactions (occurred_at)`
		_, err = db.Exec(idx)
	}
	if err == nil {
		const idxCat = `CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions (category)`
		_, err = db.Exec(idxCat)
	}
	return err
}
