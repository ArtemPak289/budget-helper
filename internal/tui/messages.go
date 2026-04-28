package tui

import (
	"budget-helper/internal/domain"
	"budget-helper/internal/service"
)

type dashboardMsg struct {
	Data service.Dashboard
}

type transactionsMsg struct {
	Rows []domain.Transaction
}

type addResultMsg struct {
	Err error
}

type budgetResultMsg struct {
	Err error
}

type exportMsg struct {
	Path string
	Err  error
}

type errMsg struct {
	Err error
}
