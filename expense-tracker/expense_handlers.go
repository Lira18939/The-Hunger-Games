package handlers

import (
	"net/http"
	"strconv"

	"expense-tracker/internal/models"
	"expense-tracker/internal/repository"
)

type ExpensesPageData struct {
	Expenses         []models.Expense
	Categories       []models.Category
	SelectedCategory int64
	From             string
	To               string
}

type ExpenseFormData struct {
	IsEdit           bool
	ID               int64
	Amount           string
	Description      string
	Date             string
	SelectedCategory int64
	Categories       []models.Category
	Error            string
}

func (h *Handler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)

	q := r.URL.Query()
	filter := repository.ExpenseFilter{
		From: q.Get("from"),
		To:   q.Get("to"),
	}
	var selectedCategory int64
	if catStr := q.Get("category"); catStr != "" {
		if id, err := strconv.ParseInt(catStr, 10, 64); err == nil {
			selectedCategory = id
			filter.CategoryID = &id
		}
	}

	expenses, err := h.expenses.GetAll(userID, filter)
	if err != nil {
		http.Error(w, "ошибка загрузки трат: "+err.Error(), http.StatusInternalServerError)
		return
	}
	categories, err := h.categories.GetAll()
	if err != nil {
		http.Error(w, "ошибка загрузки категорий: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, "expenses", ExpensesPageData{
		Expenses:         expenses,
		Categories:       categories,
		SelectedCategory: selectedCategory,
		From:             filter.From,
		To:               filter.To,
	})
}

func (h *Handler) NewExpenseForm(w http.ResponseWriter, r *http.Request) {
	h.showForm(w, r, false, 0)
}

func (h *Handler) EditExpenseForm(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Проверяем, что трата принадлежит пользователю.
	if _, err := h.expenses.GetByID(userID, id); err != nil {
		http.NotFound(w, r)
		return
	}
	h.showForm(w, r, true, id)
}

func (h *Handler) showForm(w http.ResponseWriter, r *http.Request, isEdit bool, id int64) {
	userID, _ := userIDFromContext(r)

	categories, err := h.categories.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := ExpenseFormData{
		IsEdit:     isEdit,
		ID:         id,
		Date:       today(),
		Categories: categories,
	}
	if isEdit {
		e, err := h.expenses.GetByID(userID, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		data.Amount = strconv.FormatFloat(e.Amount, 'f', 2, 64)
		data.Description = e.Description
		data.Date = e.Date
		if e.CategoryID != nil {
			data.SelectedCategory = *e.CategoryID
		}
	}
	h.render(w, "expense_form", data)
}

func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	h.saveExpense(w, r, false, 0)
}

func (h *Handler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.saveExpense(w, r, true, id)
}

func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.expenses.Delete(userID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *Handler) saveExpense(w http.ResponseWriter, r *http.Request, isEdit bool, id int64) {
	userID, _ := userIDFromContext(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "ошибка разбора формы", http.StatusBadRequest)
		return
	}
	amountStr := r.FormValue("amount")
	description := r.FormValue("description")
	date := r.FormValue("date")
	categoryStr := r.FormValue("category_id")

	var selectedCategory int64
	if categoryStr != "" {
		if cid, err := strconv.ParseInt(categoryStr, 10, 64); err == nil {
			selectedCategory = cid
		}
	}

	renderError := func(msg string) {
		categories, _ := h.categories.GetAll()
		h.renderStatus(w, "expense_form", ExpenseFormData{
			IsEdit:           isEdit,
			ID:               id,
			Amount:           amountStr,
			Description:      description,
			Date:             date,
			SelectedCategory: selectedCategory,
			Categories:       categories,
			Error:            msg,
		}, http.StatusBadRequest)
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		renderError("Сумма должна быть положительным числом")
		return
	}
	if description == "" {
		renderError("Описание не может быть пустым")
		return
	}
	if date == "" {
		renderError("Укажите дату")
		return
	}

	e := models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
	}
	if selectedCategory > 0 {
		e.CategoryID = &selectedCategory
	}

	if isEdit {
		e.ID = id
		if err := h.expenses.Update(userID, &e); err != nil {
			renderError("Ошибка сохранения: " + err.Error())
			return
		}
	} else {
		if err := h.expenses.Create(userID, &e); err != nil {
			renderError("Ошибка сохранения: " + err.Error())
			return
		}
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}