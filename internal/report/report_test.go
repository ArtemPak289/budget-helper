package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"budget-helper/internal/domain"
)

type reportCase struct {
	name   string
	list   []domain.Transaction
	income int64
	exp    int64
	net    int64
}

func TestBuildMonthlyReport(t *testing.T) {
	cases := []reportCase{
		{name: "empty", list: nil, income: 0, exp: 0, net: 0},
		{name: "mixed", list: sampleTransactions(), income: 120000, exp: -45000, net: 75000},
	}
	for _, tc := range cases {
		runReportCase(t, tc)
	}
}

func runReportCase(t *testing.T, tc reportCase) {
	report := BuildMonthlyReport(tc.list)
	if report.TotalIncome != tc.income || report.TotalExpenses != tc.exp || report.Net != tc.net {
		t.Fatalf("totals mismatch")
	}
}

func TestExportCSV(t *testing.T) {
	report := BuildMonthlyReport(sampleTransactions())
	var buf bytes.Buffer
	err := ExportCSV(&buf, report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "TOTAL_INCOME") {
		t.Fatalf("missing totals row")
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
