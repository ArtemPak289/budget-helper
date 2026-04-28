package tui

import (
	"time"

	"budget-helper/internal/config"
	"budget-helper/internal/domain"
	"budget-helper/internal/service"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/rs/zerolog"
	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	screenDashboard Screen = iota
	screenTransactions
)

type Model struct {
	svc        *service.LedgerService
	cfg        config.Config
	log        zerolog.Logger
	styles     Styles
	screen     Screen
	width      int
	height      int
	statusMsg  string
	dashboard  DashboardState
	list       TransactionsState
	addForm    AddForm
	budgetForm BudgetForm
}

type DashboardState struct {
	Data service.Dashboard
}

type TransactionsState struct {
	Viewport      viewport.Model
	Filter        domain.TransactionFilter
	MonthInput    textinput.Model
	CategoryInput textinput.Model
	SearchInput   textinput.Model
	Rows          []domain.Transaction
	Focus         FilterFocus
}

type FilterFocus int

const (
	focusNone FilterFocus = iota
	focusMonth
	focusCategory
	focusSearch
)

func NewModel(svc *service.LedgerService, cfg config.Config, log zerolog.Logger) Model {
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	monthInput := textinput.New()
	monthInput.Placeholder = "YYYY-MM"
	monthInput.SetValue(monthStart.Format("2006-01"))
	categoryInput := textinput.New()
	categoryInput.Placeholder = "Category"
	searchInput := textinput.New()
	searchInput.Placeholder = "Search"
	vp := viewport.New(0, 0)
	state := TransactionsState{
		Viewport:      vp,
		Filter:        domain.TransactionFilter{MonthStart: monthStart},
		MonthInput:    monthInput,
		CategoryInput: categoryInput,
		SearchInput:   searchInput,
	}
	return Model{
		svc:        svc,
		cfg:        cfg,
		log:        log,
		styles:     NewStyles(),
		screen:     screenDashboard,
		dashboard:  DashboardState{},
		list:       state,
		addForm:    NewAddForm(),
		budgetForm: NewBudgetForm(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadDashboardCmd(m.svc, m.list.Filter.MonthStart), loadTransactionsCmd(m.svc, m.list.Filter))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m = m.onResize(msg)
	case dashboardMsg:
		m, cmd = m.onDashboard(msg)
	case transactionsMsg:
		m, cmd = m.onTransactions(msg)
	case addResultMsg:
		m, cmd = m.onAddResult(msg)
	case budgetResultMsg:
		m, cmd = m.onBudgetResult(msg)
	case exportMsg:
		m = m.onExportResult(msg)
	case errMsg:
		m = m.onError(msg)
	case tea.KeyMsg:
		m, cmd = m.onKey(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.addForm.Visible {
		return m.viewAddForm()
	}
	if m.budgetForm.Visible {
		return m.viewBudgetForm()
	}
	if m.screen == screenTransactions {
		return m.viewTransactions()
	}
	return m.viewDashboard()
}
