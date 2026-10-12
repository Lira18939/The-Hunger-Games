package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"expense-tracker/internal/handlers"
	"expense-tracker/internal/repository"
)

func main() {

	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatal("Не удалось открыть базу данных:", err)
	}
	defer db.Close()

	applyMigrations(db)

	repo := repository.NewExpenseRepository(db)

	handlers.Init(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.ListExpenses)
	mux.HandleFunc("/add", handlers.ShowAddForm)
	mux.HandleFunc("/add/create", handlers.CreateExpense)
	mux.HandleFunc("/edit", handlers.ShowEditForm)
	mux.HandleFunc("/edit/save", handlers.UpdateExpense)
	mux.HandleFunc("/delete", handlers.DeleteExpense)

	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func applyMigrations(db *sql.DB) {
	sqlBytes, err := os.ReadFile("migrations/001_create_expenses.sql")
	if err != nil {
		log.Fatal("Не удалось прочитать миграцию:", err)
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		log.Fatal("Не удалось выполнить миграцию:", err)
	}
}