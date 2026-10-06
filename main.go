package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"expense-tracker/internal/handlers"
	"expense-tracker/internal/repository"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "expenses.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("не удалось открыть БД: %v", err)
	}
	defer db.Close()

	// SQLite не любит параллельные писатели — одного соединения достаточно.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatalf("не удалось включить foreign_keys: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("ошибка миграций: %v", err)
	}

	expenseRepo := repository.NewExpenseRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	h, err := handlers.New(expenseRepo, categoryRepo)
	if err != nil {
		log.Fatalf("ошибка инициализации обработчиков: %v", err)
	}

	mux := http.NewServeMux()

	// Статика
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Главная
	mux.HandleFunc("GET /{$}", h.Home)

	// Траты
	mux.HandleFunc("GET /expenses", h.ListExpenses)
	mux.HandleFunc("GET /expenses/new", h.NewExpenseForm)
	mux.HandleFunc("POST /expenses", h.CreateExpense)
	mux.HandleFunc("GET /expenses/{id}/edit", h.EditExpenseForm)
	mux.HandleFunc("POST /expenses/{id}", h.UpdateExpense)
	mux.HandleFunc("POST /expenses/{id}/delete", h.DeleteExpense)

	// Категории
	mux.HandleFunc("GET /categories", h.ListCategories)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Printf("Сервер запущен на http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("ошибка сервера: %v", err)
	}
}

func runMigrations(db *sql.DB) error {
	data, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(data))
	return err
}