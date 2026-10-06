package repository

import (
	"database/sql"
	"strings"

	"expense-tracker/internal/models"
)

// ExpenseFilter описывает возможные фильтры списка трат.
// Все поля опциональны: если CategoryID == nil, From == "", To == "" —
// фильтр не применяется.
type ExpenseFilter struct {
	CategoryID *int64
	From       string // YYYY-MM-DD
	To         string // YYYY-MM-DD
}

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

const expenseSelect = `
	SELECT e.id, e.amount, e.description, e.date,
	       e.category_id, COALESCE(c.name, '')
	FROM expenses e
	LEFT JOIN categories c ON c.id = e.category_id
`

func (r *ExpenseRepository) Create(e *models.Expense) error {
	_, err := r.db.Exec(
		`INSERT INTO expenses (amount, description, date, category_id)
		 VALUES (?, ?, ?, ?)`,
		e.Amount, e.Description, e.Date, e.CategoryID,
	)
	return err
}

func (r *ExpenseRepository) Update(e *models.Expense) error {
	_, err := r.db.Exec(
		`UPDATE expenses
		 SET amount = ?, description = ?, date = ?, category_id = ?
		 WHERE id = ?`,
		e.Amount, e.Description, e.Date, e.CategoryID, e.ID,
	)
	return err
}

func (r *ExpenseRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM expenses WHERE id = ?", id)
	return err
}

func (r *ExpenseRepository) GetByID(id int64) (*models.Expense, error) {
	row := r.db.QueryRow(expenseSelect+" WHERE e.id = ?", id)
	var e models.Expense
	err := row.Scan(&e.ID, &e.Amount, &e.Description, &e.Date,
		&e.CategoryID, &e.CategoryName)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetAll возвращает траты с учётом фильтров.
// SQL собирается динамически, но значения всегда передаются через placeholder'ы —
// это защищает от SQL-инъекций.
func (r *ExpenseRepository) GetAll(f ExpenseFilter) ([]models.Expense, error) {
	var sb strings.Builder
	sb.WriteString(expenseSelect)
	sb.WriteString(" WHERE 1=1")

	var args []any

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
		if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &e.Date,
			&e.CategoryID, &e.CategoryName); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}