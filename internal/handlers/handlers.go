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

func Init(r *repository.ExpenseRepository) {
	repo = r
}

func ListExpenses(w http.ResponseWriter, r *http.Request) {
	expenses, err := repo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка получения трат: "+err.Error(), http.StatusInternalServerError)
		return
	}
	templates.ExecuteTemplate(w, "index.html", expenses)
}

func ShowAddForm(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "add.html", nil)
}

func CreateExpense(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Неверная сумма", http.StatusBadRequest)
		return
	}

	e := models.Expense{
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
	}
	if err := repo.Create(e); err != nil {
		http.Error(w, "Ошибка сохранения: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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
	templates.ExecuteTemplate(w, "edit.html", e)
}

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

	e := models.Expense{
		ID:          id,
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
	}
	if err := repo.Update(e); err != nil {
		http.Error(w, "Ошибка обновления: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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