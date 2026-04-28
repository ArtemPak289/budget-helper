package tui

import (
	"strconv"
	"strings"
	"time"

	"budget-helper/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) onResize(msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height
	m.list.Viewport.Width = msg.Width
	m.list.Viewport.Height = msg.Height - 8
	return m
}

func (m Model) onDashboard(msg dashboardMsg) (Model, tea.Cmd) {
	m.dashboard.Data = msg.Data
	m.statusMsg = ""
	return m, nil
}

func (m Model) onTransactions(msg transactionsMsg) (Model, tea.Cmd) {
	m.list.Rows = msg.Rows
	m.list.Viewport.SetContent(renderTransactions(msg.Rows))
	m.statusMsg = ""
	return m, nil
}

func (m Model) onAddResult(msg addResultMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if msg.Err != nil {
		m.addForm.Error = msg.Err.Error()
	}
	if msg.Err == nil {
		m.addForm = NewAddForm()
		m.addForm.Visible = false
		m.statusMsg = "transaction added"
		cmd = reloadCmds(m)
	}
	return m, cmd
}

func (m Model) onBudgetResult(msg budgetResultMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if msg.Err != nil {
		m.budgetForm.Error = msg.Err.Error()
	}
	if msg.Err == nil {
		m.budgetForm = NewBudgetForm()
		m.budgetForm.Visible = false
		m.statusMsg = "budget saved"
		cmd = reloadCmds(m)
	}
	return m, cmd
}

func (m Model) onExportResult(msg exportMsg) Model {
	if msg.Err != nil {
		m.statusMsg = "export failed: " + msg.Err.Error()
	}
	if msg.Err == nil {
		m.statusMsg = "exported to " + msg.Path
	}
	return m
}

func (m Model) onError(msg errMsg) Model {
	if msg.Err != nil {
		m.statusMsg = msg.Err.Error()
	}
	return m
}

func (m Model) onKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.addForm.Visible {
		m, cmd = m.onAddKey(msg)
	}
	if !m.addForm.Visible && m.budgetForm.Visible {
		m, cmd = m.onBudgetKey(msg)
	}
	if !m.addForm.Visible && !m.budgetForm.Visible {
		if m.screen == screenTransactions {
			m, cmd = m.onTransactionsKey(msg)
		}
		if m.screen == screenDashboard {
			m, cmd = m.onDashboardKey(msg)
		}
	}
	return m, cmd
}

func (m Model) onDashboardKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "t":
		m.screen = screenTransactions
	case "a":
		m = m.openAddForm()
	case "b":
		m = m.openBudgetForm()
	case "e":
		cmd = exportReportCmd(m.svc, m.list.Filter.MonthStart, m.cfg.ExportDir)
	case "q", "ctrl+c":
		cmd = tea.Quit
	}
	return m, cmd
}

func (m Model) onTransactionsKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list.Viewport, _ = m.list.Viewport.Update(msg)
	switch msg.String() {
	case "d":
		m.screen = screenDashboard
		cmd = loadDashboardCmd(m.svc, m.list.Filter.MonthStart)
	case "a":
		m = m.openAddForm()
	case "b":
		m = m.openBudgetForm()
	case "e":
		cmd = exportReportCmd(m.svc, m.list.Filter.MonthStart, m.cfg.ExportDir)
	case "m":
		m = m.setFilterFocus(focusMonth)
	case "c":
		m = m.setFilterFocus(focusCategory)
	case "s":
		m = m.setFilterFocus(focusSearch)
	case "enter":
		m, cmd = m.applyFilters()
	case "esc":
		m = m.setFilterFocus(focusNone)
	case "q", "ctrl+c":
		cmd = tea.Quit
	}
	m = m.updateFilterInputs(msg)
	return m, cmd
}

func (m Model) onAddKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	key := msg.String()
	if key == "esc" {
		m.addForm = NewAddForm()
		m.addForm.Visible = false
	}
	if key == "tab" {
		m = m.nextAddFocus()
	}
	if key == "enter" {
		m, cmd = m.submitAdd()
	}
	m = m.updateAddInputs(msg)
	return m, cmd
}

func (m Model) onBudgetKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	key := msg.String()
	if key == "esc" {
		m.budgetForm = NewBudgetForm()
		m.budgetForm.Visible = false
	}
	if key == "tab" {
		m = m.nextBudgetFocus()
	}
	if key == "enter" {
		m, cmd = m.submitBudget()
	}
	m = m.updateBudgetInputs(msg)
	return m, cmd
}

func (m Model) updateFilterInputs(msg tea.Msg) Model {
	var cmd tea.Cmd
	m.list.MonthInput, cmd = m.list.MonthInput.Update(msg)
	_ = cmd
	m.list.CategoryInput, cmd = m.list.CategoryInput.Update(msg)
	_ = cmd
	m.list.SearchInput, cmd = m.list.SearchInput.Update(msg)
	_ = cmd
	return m
}

func (m Model) applyFilters() (Model, tea.Cmd) {
	var cmd tea.Cmd
	monthStart, err := parseMonthInput(m.list.MonthInput.Value())
	if err != nil {
		m.statusMsg = err.Error()
	}
	if err == nil {
		m.list.Filter.MonthStart = monthStart
		m.list.Filter.Category = strings.TrimSpace(m.list.CategoryInput.Value())
		m.list.Filter.Search = strings.TrimSpace(m.list.SearchInput.Value())
		cmd = loadTransactionsCmd(m.svc, m.list.Filter)
	}
	return m, cmd
}

func (m Model) updateAddInputs(msg tea.Msg) Model {
	var cmd tea.Cmd
	m.addForm.Amount, cmd = m.addForm.Amount.Update(msg)
	_ = cmd
	m.addForm.Currency, cmd = m.addForm.Currency.Update(msg)
	_ = cmd
	m.addForm.Category, cmd = m.addForm.Category.Update(msg)
	_ = cmd
	m.addForm.Description, cmd = m.addForm.Description.Update(msg)
	_ = cmd
	return m
}

func (m Model) updateBudgetInputs(msg tea.Msg) Model {
	var cmd tea.Cmd
	m.budgetForm.Category, cmd = m.budgetForm.Category.Update(msg)
	_ = cmd
	m.budgetForm.Currency, cmd = m.budgetForm.Currency.Update(msg)
	_ = cmd
	m.budgetForm.Amount, cmd = m.budgetForm.Amount.Update(msg)
	_ = cmd
	return m
}

func (m Model) nextAddFocus() Model {
	m.addForm.Focus = (m.addForm.Focus + 1) % 4
	m = m.setAddFocus()
	return m
}

func (m Model) nextBudgetFocus() Model {
	m.budgetForm.Focus = (m.budgetForm.Focus + 1) % 3
	m = m.setBudgetFocus()
	return m
}

func (m Model) setFilterFocus(focus FilterFocus) Model {
	m.list.Focus = focus
	m.list.MonthInput.Blur()
	m.list.CategoryInput.Blur()
	m.list.SearchInput.Blur()
	if focus == focusMonth {
		m.list.MonthInput.Focus()
	}
	if focus == focusCategory {
		m.list.CategoryInput.Focus()
	}
	if focus == focusSearch {
		m.list.SearchInput.Focus()
	}
	return m
}

func (m Model) setAddFocus() Model {
	inputs := m.addForm.Inputs()
	for i := range inputs {
		inputs[i].Blur()
		if i == m.addForm.Focus {
			inputs[i].Focus()
		}
	}
	m.addForm.Amount = inputs[0]
	m.addForm.Currency = inputs[1]
	m.addForm.Category = inputs[2]
	m.addForm.Description = inputs[3]
	return m
}

func (m Model) setBudgetFocus() Model {
	inputs := m.budgetForm.Inputs()
	for i := range inputs {
		inputs[i].Blur()
		if i == m.budgetForm.Focus {
			inputs[i].Focus()
		}
	}
	m.budgetForm.Category = inputs[0]
	m.budgetForm.Currency = inputs[1]
	m.budgetForm.Amount = inputs[2]
	return m
}

func (m Model) openAddForm() Model {
	m.addForm = NewAddForm()
	m.addForm.Visible = true
	m.addForm.Focus = 0
	m = m.setAddFocus()
	return m
}

func (m Model) openBudgetForm() Model {
	m.budgetForm = NewBudgetForm()
	m.budgetForm.Visible = true
	m.budgetForm.Focus = 0
	m = m.setBudgetFocus()
	return m
}

func (m Model) submitAdd() (Model, tea.Cmd) {
	var cmd tea.Cmd
	amount, err := strconv.ParseInt(strings.TrimSpace(m.addForm.Amount.Value()), 10, 64)
	if err != nil {
		m.addForm.Error = "invalid amount"
	}
	if err == nil {
		currency := strings.TrimSpace(m.addForm.Currency.Value())
		category := strings.TrimSpace(m.addForm.Category.Value())
		desc := strings.TrimSpace(m.addForm.Description.Value())
		cmd = addTransactionCmd(m.svc, amount, currency, category, desc)
	}
	return m, cmd
}

func (m Model) submitBudget() (Model, tea.Cmd) {
	var cmd tea.Cmd
	amount, err := strconv.ParseInt(strings.TrimSpace(m.budgetForm.Amount.Value()), 10, 64)
	if err != nil {
		m.budgetForm.Error = "invalid amount"
	}
	if err == nil {
		category := strings.TrimSpace(m.budgetForm.Category.Value())
		currency := strings.TrimSpace(m.budgetForm.Currency.Value())
		cmd = setBudgetCmd(m.svc, m.list.Filter.MonthStart, category, currency, amount)
	}
	return m, cmd
}

func reloadCmds(m Model) tea.Cmd {
	return tea.Batch(loadDashboardCmd(m.svc, m.list.Filter.MonthStart), loadTransactionsCmd(m.svc, m.list.Filter))
}

func renderTransactions(rows []domain.Transaction) string {
	lines := make([]string, 0, len(rows))
	for _, tx := range rows {
		lines = append(lines, formatTxLine(tx))
	}
	return strings.Join(lines, "\n")
}

func formatTxLine(tx domain.Transaction) string {
	date := tx.OccurredAt.Format("2006-01-02")
	amount := strconv.FormatInt(tx.AmountMinor, 10)
	return date + " | " + tx.Category + " | " + tx.Description + " | " + amount + " " + tx.Currency
}

func parseMonthInput(input string) (time.Time, error) {
	var (
		val time.Time
		err error
	)
	val, err = time.Parse("2006-01", strings.TrimSpace(input))
	return time.Date(val.Year(), val.Month(), 1, 0, 0, 0, 0, time.Local), err
}
