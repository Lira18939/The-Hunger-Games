package models

// Category — категория трат.
type Category struct {
	ID   int64
	Name string
}

// Expense — одна запись о трате.
// CategoryID — указатель, потому что категория может быть не выбрана (NULL).
type Expense struct {
	ID           int64
	Amount       float64
	Description  string
	Date         string // YYYY-MM-DD
	CategoryID   *int64
	CategoryName string // подтягивается через JOIN
}