package repository

import (
	"database/sql"

	"expense-tracker/internal/models"
)

// ExpenseRepository — слой доступа к таблице expenses.
type ExpenseRepository struct {
	db *sql.DB
}

// NewExpenseRepository создаёт репозиторий.
func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// Create — добавить трату.
func (r *ExpenseRepository) Create(e models.Expense) error {
	_, err := r.db.Exec(
		"INSERT INTO expenses (amount, description, date, category_id) VALUES (?, ?, ?, ?)",
		e.Amount, e.Description, e.Date, e.CategoryID,
	)
	return err
}

// GetByID — получить одну трату.
func (r *ExpenseRepository) GetByID(id int) (models.Expense, error) {
	var e models.Expense
	err := r.db.QueryRow(`
		SELECT e.id, e.amount, e.description, e.date,
		       COALESCE(e.category_id, 0),
		       COALESCE(c.name, '')
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE e.id = ?`, id,
	).Scan(&e.ID, &e.Amount, &e.Description, &e.Date, &e.CategoryID, &e.CategoryName)
	return e, err
}

// Update — обновить трату.
func (r *ExpenseRepository) Update(e models.Expense) error {
	_, err := r.db.Exec(
		"UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ? WHERE id = ?",
		e.Amount, e.Description, e.Date, e.CategoryID, e.ID,
	)
	return err
}

// Delete — удалить трату.
func (r *ExpenseRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM expenses WHERE id = ?", id)
	return err
}

// GetFiltered — получить траты с фильтрами.
// categoryID = 0 означает «не фильтровать по категории».
// dateFrom = "" или dateTo = "" означает «не фильтровать по дате».
func (r *ExpenseRepository) GetFiltered(categoryID int, dateFrom, dateTo string) ([]models.Expense, error) {
	query := `
		SELECT e.id, e.amount, e.description, e.date,
		       COALESCE(e.category_id, 0),
		       COALESCE(c.name, '')
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE 1 = 1`

	var args []interface{}

	if categoryID > 0 {
		query += " AND e.category_id = ?"
		args = append(args, categoryID)
	}
	if dateFrom != "" {
		query += " AND e.date >= ?"
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		query += " AND e.date <= ?"
		args = append(args, dateTo)
	}
	query += " ORDER BY e.id DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &e.Date, &e.CategoryID, &e.CategoryName); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

// GetAllCategories — получить все категории.
func (r *ExpenseRepository) GetAllCategories() ([]models.Category, error) {
	rows, err := r.db.Query("SELECT id, name FROM categories ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}