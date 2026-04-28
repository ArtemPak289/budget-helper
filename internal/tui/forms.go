package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
)

type AddForm struct {
	Visible     bool
	Focus       int
	Amount      textinput.Model
	Currency    textinput.Model
	Category    textinput.Model
	Description textinput.Model
	Error       string
}

type BudgetForm struct {
	Visible  bool
	Focus    int
	Category textinput.Model
	Currency textinput.Model
	Amount   textinput.Model
	Error    string
}

func NewAddForm() AddForm {
	amount := textinput.New()
	amount.Placeholder = "-45000"
	currency := textinput.New()
	currency.Placeholder = "USD"
	category := textinput.New()
	category.Placeholder = "Food"
	desc := textinput.New()
	desc.Placeholder = "Description"
	return AddForm{Amount: amount, Currency: currency, Category: category, Description: desc}
}

func NewBudgetForm() BudgetForm {
	category := textinput.New()
	category.Placeholder = "Food"
	currency := textinput.New()
	currency.Placeholder = "USD"
	amount := textinput.New()
	amount.Placeholder = "100000"
	return BudgetForm{Category: category, Currency: currency, Amount: amount}
}

func (f AddForm) Inputs() []textinput.Model {
	return []textinput.Model{f.Amount, f.Currency, f.Category, f.Description}
}

func (f BudgetForm) Inputs() []textinput.Model {
	return []textinput.Model{f.Category, f.Currency, f.Amount}
}
