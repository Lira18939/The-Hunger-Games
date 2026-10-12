package main

import (
	"fmt"
	"net/http"

	"expense-tracker/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	// Страница со списком трат
	mux.HandleFunc("/", handlers.ListExpenses)

	// Страница с формой добавления
	mux.HandleFunc("/add", handlers.ShowAddForm)

	// Обработка отправки формы
	mux.HandleFunc("/add/create", handlers.CreateExpense)

	// Раздача статических файлов (CSS)
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}