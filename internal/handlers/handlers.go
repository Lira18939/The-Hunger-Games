package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"expense-tracker/internal/models"
)

var expenses []models.Expense

var nextID = 1

var templates = template.Must(template.ParseGlob("web/templates/*.html"))

func ListExpenses(w http.ResponseWriter, r *http.Request) {
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
	description := r.FormValue("description")
	date := r.FormValue("date")
	
	expense := models.Expense{
		ID:          nextID,
		Amount:      amount,
		Description: description,
		Date:        date,
	}
	nextID++
	expenses = append(expenses, expense)
	
	http.Redirect(w, r, "/", http.StatusSeeOther)
}