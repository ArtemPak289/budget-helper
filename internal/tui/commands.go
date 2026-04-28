package tui

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"budget-helper/internal/domain"
	"budget-helper/internal/service"
	tea "github.com/charmbracelet/bubbletea"
)

func loadDashboardCmd(svc *service.LedgerService, monthStart time.Time) tea.Cmd {
	return func() tea.Msg {
		var out tea.Msg
		data, err := svc.Dashboard(context.Background(), monthStart.Year(), monthStart.Month())
		if err != nil {
			out = errMsg{Err: err}
		}
		if err == nil {
			out = dashboardMsg{Data: data}
		}
		return out
	}
}

func loadTransactionsCmd(svc *service.LedgerService, filter domain.TransactionFilter) tea.Cmd {
	return func() tea.Msg {
		var out tea.Msg
		list, err := svc.ListTransactions(context.Background(), filter)
		if err != nil {
			out = errMsg{Err: err}
		}
		if err == nil {
			out = transactionsMsg{Rows: list}
		}
		return out
	}
}

func addTransactionCmd(svc *service.LedgerService, amount int64, currency, category, description string) tea.Cmd {
	return func() tea.Msg {
		var out tea.Msg
		_, err := svc.AddTransaction(context.Background(), amount, currency, category, description, time.Time{})
		out = addResultMsg{Err: err}
		return out
	}
}

func setBudgetCmd(svc *service.LedgerService, monthStart time.Time, category, currency string, amount int64) tea.Cmd {
	return func() tea.Msg {
		var out tea.Msg
		_, err := svc.SetMonthlyBudget(context.Background(), monthStart.Year(), monthStart.Month(), category, currency, amount)
		out = budgetResultMsg{Err: err}
		return out
	}
}

func exportReportCmd(svc *service.LedgerService, monthStart time.Time, dir string) tea.Cmd {
	return func() tea.Msg {
		var out tea.Msg
		path, err := writeReport(svc, monthStart, dir)
		out = exportMsg{Path: path, Err: err}
		return out
	}
}

func writeReport(svc *service.LedgerService, monthStart time.Time, dir string) (string, error) {
	var (
		path string
		err  error
	)
	if dir == "" {
		dir = "."
	}
	path = filepath.Join(dir, "report-"+monthStart.Format("2006-01")+".csv")
	err = os.MkdirAll(dir, 0o755)
	if err == nil {
		var file *os.File
		file, err = os.Create(path)
		if err == nil {
			defer file.Close()
			err = svc.ExportMonthlyReport(context.Background(), monthStart.Year(), monthStart.Month(), file)
		}
	}
	return path, err
}
