package models

// Expense — структура одной траты
type Expense struct {
	ID          int
	Amount      float64
	Description string
	Date        string
}