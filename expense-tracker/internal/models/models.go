package models

type Category struct {
	ID   int64
	Name string
}

type Expense struct {
	ID           int64
	UserID       int64
	Amount       float64
	Description  string
	Date         string
	CategoryID   *int64
	CategoryName string
}

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type Session struct {
	SessionID string
	UserID    int64
	ExpiresAt string // RFC3339
}