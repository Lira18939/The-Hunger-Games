package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"expense-tracker/internal/models"
	"expense-tracker/internal/repository"
)

var templates = template.Must(template.ParseGlob("web/templates/*.html"))

var repo *repository.ExpenseRepository

// Init — вызывается из main, передаёт репозиторий обработчикам.
func Init(r *repository.ExpenseRepository) {
	repo = r
}

// filterData — данные, передаваемые в шаблон index.html.
type filterData struct {
	Expenses   []models.Expense
	Categories []models.Category
	CategoryID int
	DateFrom   string
	DateTo     string
}

// ListExpenses — список трат с фильтрами.
func ListExpenses(w http.ResponseWriter, r *http.Request) {
	categoryID, _ := strconv.Atoi(r.URL.Query().Get("category"))
	dateFrom := r.URL.Query().Get("from")
	dateTo := r.URL.Query().Get("to")

	expenses, err := repo.GetFiltered(categoryID, dateFrom, dateTo)
	if err != nil {
		http.Error(w, "Ошибка получения трат: "+err.Error(), http.StatusInternalServerError)
		return
	}

	categories, err := repo.GetAllCategories()
	if err != nil {
		http.Error(w, "Ошибка получения категорий: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := filterData{
		Expenses:   expenses,
		Categories: categories,
		CategoryID: categoryID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	}
	templates.ExecuteTemplate(w, "index.html", data)
}

// ShowAddForm — форма добавления.
func ShowAddForm(w http.ResponseWriter, r *http.Request) {
	categories, err := repo.GetAllCategories()
	if err != nil {
		http.Error(w, "Ошибка категорий", http.StatusInternalServerError)
		return
	}
	data := struct {
		Categories []models.Category
	}{Categories: categories}
	templates.ExecuteTemplate(w, "add.html", data)
}

// CreateExpense — обработка отправки формы добавления.
func CreateExpense(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Неверная сумма", http.StatusBadRequest)
		return
	}
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))

	e := models.Expense{
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
		CategoryID:  categoryID,
	}
	if err := repo.Create(e); err != nil {
		http.Error(w, "Ошибка сохранения: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ShowEditForm — форма редактирования.
func ShowEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Неверный id", http.StatusBadRequest)
		return
	}
	e, err := repo.GetByID(id)
	if err != nil {
		http.Error(w, "Трата не найдена", http.StatusNotFound)
		return
	}
	categories, err := repo.GetAllCategories()
	if err != nil {
		http.Error(w, "Ошибка категорий", http.StatusInternalServerError)
		return
	}
	data := struct {
		Expense    models.Expense
		Categories []models.Category
	}{Expense: e, Categories: categories}
	templates.ExecuteTemplate(w, "edit.html", data)
}

// UpdateExpense — обработка формы редактирования.
func UpdateExpense(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Неверный id", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Неверная сумма", http.StatusBadRequest)
		return
	}
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))

	e := models.Expense{
		ID:          id,
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
		CategoryID:  categoryID,
	}
	if err := repo.Update(e); err != nil {
		http.Error(w, "Ошибка обновления: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteExpense — удаление траты.
func DeleteExpense(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Неверный id", http.StatusBadRequest)
		return
	}
	if err := repo.Delete(id); err != nil {
		http.Error(w, "Ошибка удаления: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}