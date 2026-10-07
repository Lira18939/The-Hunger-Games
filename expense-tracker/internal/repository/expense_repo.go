package repository

import (
	"database/sql"
	"strings"

	"expense-tracker/internal/models"
)

type ExpenseFilter struct {
	CategoryID *int64
	From       string
	To         string
}

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

const expenseSelect = `
	SELECT e.id, e.user_id, e.amount, e.description, e.date,
	       e.category_id, COALESCE(c.name, '')
	FROM expenses e
	LEFT JOIN categories c ON c.id = e.category_id
`

func (r *ExpenseRepository) Create(userID int64, e *models.Expense) error {
	_, err := r.db.Exec(
		`INSERT INTO expenses (user_id, amount, description, date, category_id)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, e.Amount, e.Description, e.Date, e.CategoryID,
	)
	return err
}

// Update — обновляем только если трата принадлежит этому пользователю.
func (r *ExpenseRepository) Update(userID int64, e *models.Expense) error {
	_, err := r.db.Exec(
		`UPDATE expenses
		 SET amount = ?, description = ?, date = ?, category_id = ?
		 WHERE id = ? AND user_id = ?`,
		e.Amount, e.Description, e.Date, e.CategoryID, e.ID, userID,
	)
	return err
}

func (r *ExpenseRepository) Delete(userID, id int64) error {
	_, err := r.db.Exec(
		"DELETE FROM expenses WHERE id = ? AND user_id = ?",
		id, userID,
	)
	return err
}

func (r *ExpenseRepository) GetByID(userID, id int64) (*models.Expense, error) {
	row := r.db.QueryRow(
		expenseSelect+" WHERE e.id = ? AND e.user_id = ?",
		id, userID,
	)
	var e models.Expense
	err := row.Scan(&e.ID, &e.UserID, &e.Amount, &e.Description, &e.Date,
		&e.CategoryID, &e.CategoryName)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExpenseRepository) GetAll(userID int64, f ExpenseFilter) ([]models.Expense, error) {
	var sb strings.Builder
	sb.WriteString(expenseSelect)
	sb.WriteString(" WHERE e.user_id = ?")
	args := []any{userID}

	if f.CategoryID != nil {
		sb.WriteString(" AND e.category_id = ?")
		args = append(args, *f.CategoryID)
	}
	if f.From != "" {
		sb.WriteString(" AND e.date >= ?")
		args = append(args, f.From)
	}
	if f.To != "" {
		sb.WriteString(" AND e.date <= ?")
		args = append(args, f.To)
	}
	sb.WriteString(" ORDER BY e.date DESC, e.id DESC")

	rows, err := r.db.Query(sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.UserID, &e.Amount, &e.Description, &e.Date,
			&e.CategoryID, &e.CategoryName); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}