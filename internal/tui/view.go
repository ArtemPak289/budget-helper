package tui

import (
	"fmt"
	"strings"
)

func (m Model) viewDashboard() string {
	header := m.styles.Title.Render("Ledger Dashboard")
	stats := m.styles.Border.Render(m.renderStats())
	cats := m.styles.Border.Render(m.renderTopCategories())
	bars := m.styles.Border.Render(m.renderBars())
	budgets := m.styles.Border.Render(m.renderBudgets())
	help := m.renderHelp("t=transactions a=add b=budget e=export q=quit")
	status := m.renderStatus()
	return strings.Join([]string{header, stats, cats, bars, budgets, help, status}, "\n")
}

func (m Model) viewTransactions() string {
	header := m.styles.Title.Render("Transactions")
	filters := m.styles.Border.Render(m.renderFilters())
	list := m.styles.Border.Render(m.list.Viewport.View())
	help := m.renderHelp("d=dashboard m=month c=category s=search enter=apply a=add b=budget e=export q=quit")
	status := m.renderStatus()
	return strings.Join([]string{header, filters, list, help, status}, "\n")
}

func (m Model) viewAddForm() string {
	header := m.styles.Title.Render("Add Transaction")
	body := m.styles.Border.Render(m.renderAddForm())
	help := m.renderHelp("tab=next enter=submit esc=cancel")
	status := m.renderStatus()
	return strings.Join([]string{header, body, help, status}, "\n")
}

func (m Model) viewBudgetForm() string {
	header := m.styles.Title.Render("Set Monthly Budget")
	body := m.styles.Border.Render(m.renderBudgetForm())
	help := m.renderHelp("tab=next enter=submit esc=cancel")
	status := m.renderStatus()
	return strings.Join([]string{header, body, help, status}, "\n")
}

func (m Model) renderStats() string {
	data := m.dashboard.Data
	lines := []string{
		fmt.Sprintf("Current balance: %d %s", data.Net, data.Currency),
		fmt.Sprintf("This month income: %d %s", data.TotalIncome, data.Currency),
		fmt.Sprintf("This month expenses: %d %s", data.TotalExpenses, data.Currency),
		fmt.Sprintf("Net result: %d %s", data.Net, data.Currency),
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderTopCategories() string {
	lines := []string{"Top 5 categories"}
	if len(m.dashboard.Data.TopCategories) == 0 {
		lines = append(lines, "None")
	}
	for i, item := range m.dashboard.Data.TopCategories {
		lines = append(lines, fmt.Sprintf("%d. %s: %d", i+1, item.Category, item.AmountMinor))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderBars() string {
	lines := []string{"Expense chart"}
	if len(m.dashboard.Data.ExpenseBars) == 0 {
		lines = append(lines, "None")
	}
	for _, bar := range m.dashboard.Data.ExpenseBars {
		lines = append(lines, fmt.Sprintf("%s | %s", bar.Category, bar.Bar))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderBudgets() string {
	lines := []string{"Budgets"}
	if len(m.dashboard.Data.Budgets) == 0 {
		lines = append(lines, "None")
	}
	for _, item := range m.dashboard.Data.Budgets {
		line := fmt.Sprintf("%s: %d / %d", item.Category, item.SpentMinor, item.BudgetMinor)
		if item.Exceeded {
			line = m.styles.Warning.Render(line + " (exceeded)")
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderFilters() string {
	month := m.list.MonthInput.View()
	cat := m.list.CategoryInput.View()
	search := m.list.SearchInput.View()
	lines := []string{
		"Month: " + month,
		"Category: " + cat,
		"Search: " + search,
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderAddForm() string {
	lines := []string{
		"Amount: " + m.addForm.Amount.View(),
		"Currency: " + m.addForm.Currency.View(),
		"Category: " + m.addForm.Category.View(),
		"Description: " + m.addForm.Description.View(),
	}
	if m.addForm.Error != "" {
		lines = append(lines, m.styles.Warning.Render(m.addForm.Error))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderBudgetForm() string {
	lines := []string{
		"Category: " + m.budgetForm.Category.View(),
		"Currency: " + m.budgetForm.Currency.View(),
		"Amount: " + m.budgetForm.Amount.View(),
	}
	if m.budgetForm.Error != "" {
		lines = append(lines, m.styles.Warning.Render(m.budgetForm.Error))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderHelp(text string) string {
	return m.styles.Help.Render(text)
}

func (m Model) renderStatus() string {
	out := ""
	if m.statusMsg != "" {
		out = m.styles.Status.Render(m.statusMsg)
	}
	return out
}
