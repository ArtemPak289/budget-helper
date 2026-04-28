package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"

	"budget-helper/internal/domain"
)

type MonthlyReport struct {
	Currency      string
	TotalIncome   int64
	TotalExpenses int64
	Net           int64
	TopCategories []domain.CategoryTotal
}

type ExpenseBar struct {
	Category    string
	AmountMinor int64
	Bar         string
}

func BuildMonthlyReport(list []domain.Transaction) MonthlyReport {
	report := MonthlyReport{}
	report.Currency = pickCurrency(list)
	report.TotalIncome, report.TotalExpenses = sumTotals(list)
	report.Net = report.TotalIncome + report.TotalExpenses
	report.TopCategories = TopExpenseCategories(list, 5)
	return report
}

func TopExpenseCategories(list []domain.Transaction, limit int) []domain.CategoryTotal {
	cats := map[string]int64{}
	for _, tx := range list {
		if tx.AmountMinor < 0 {
			cats[tx.Category] += tx.AmountMinor
		}
	}
	return sortCategories(cats, limit)
}

func BuildExpenseBars(list []domain.CategoryTotal, width int) []ExpenseBar {
	bars := make([]ExpenseBar, 0, len(list))
	max := maxAbs(list)
	for _, item := range list {
		bars = append(bars, ExpenseBar{Category: item.Category, AmountMinor: item.AmountMinor, Bar: barFor(item.AmountMinor, max, width)})
	}
	return bars
}

func ExportCSV(w io.Writer, report MonthlyReport) error {
	var err error
	writer := csv.NewWriter(w)
	err = writer.Write([]string{"label", "amount_minor"})
	if err == nil {
		for _, item := range report.TopCategories {
			err = writer.Write([]string{fmt.Sprintf("CATEGORY:%s", item.Category), fmt.Sprintf("%d", item.AmountMinor)})
			if err != nil {
				break
			}
		}
	}
	if err == nil {
		err = writeTotals(writer, report)
	}
	writer.Flush()
	if err == nil {
		err = writer.Error()
	}
	return err
}

func writeTotals(writer *csv.Writer, report MonthlyReport) error {
	var err error
	rows := [][]string{
		{"TOTAL_INCOME", fmt.Sprintf("%d", report.TotalIncome)},
		{"TOTAL_EXPENSES", fmt.Sprintf("%d", report.TotalExpenses)},
		{"NET", fmt.Sprintf("%d", report.Net)},
	}
	for _, row := range rows {
		err = writer.Write(row)
		if err != nil {
			break
		}
	}
	return err
}

func pickCurrency(list []domain.Transaction) string {
	currency := ""
	for _, tx := range list {
		if tx.Currency != "" {
			currency = tx.Currency
			break
		}
	}
	return currency
}

func sumTotals(list []domain.Transaction) (int64, int64) {
	var income int64
	var expenses int64
	for _, tx := range list {
		if tx.AmountMinor >= 0 {
			income += tx.AmountMinor
		} else {
			expenses += tx.AmountMinor
		}
	}
	return income, expenses
}

func sortCategories(cats map[string]int64, limit int) []domain.CategoryTotal {
	result := make([]domain.CategoryTotal, 0, len(cats))
	for cat, amount := range cats {
		result = append(result, domain.CategoryTotal{Category: cat, AmountMinor: amount})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].AmountMinor == result[j].AmountMinor {
			return result[i].Category < result[j].Category
		}
		return result[i].AmountMinor < result[j].AmountMinor
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func maxAbs(list []domain.CategoryTotal) int64 {
	var max int64
	for _, item := range list {
		val := item.AmountMinor
		if val < 0 {
			val = -val
		}
		if val > max {
			max = val
		}
	}
	return max
}

func barFor(amount, max int64, width int) string {
	out := ""
	if max != 0 && width > 0 {
		val := amount
		if val < 0 {
			val = -val
		}
		size := int((float64(val) / float64(max)) * float64(width))
		if size == 0 && val > 0 {
			size = 1
		}
		out = repeatChar('#', size)
	}
	return out
}

func repeatChar(ch rune, count int) string {
	out := ""
	if count > 0 {
		buf := make([]rune, 0, count)
		for i := 0; i < count; i++ {
			buf = append(buf, ch)
		}
		out = string(buf)
	}
	return out
}
