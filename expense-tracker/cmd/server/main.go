package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatalf("не удалось включить foreign_keys: %v", err)
	}
	if err := runMigrations(db, "migrations"); err != nil {
		log.Fatalf("ошибка миграций: %v", err)
	}

	expenseRepo := repository.NewExpenseRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Подчистим истекшие сессии при старте.
	_ = sessionRepo.DeleteExpired()

	h, err := handlers.New(expenseRepo, categoryRepo, userRepo, sessionRepo)
	if err != nil {
		log.Fatalf("ошибка инициализации обработчиков: %v", err)
	}

	mux := http.NewServeMux()

	// Статика и главная — публичные.
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("GET /{$}", h.Home)

	// Авторизация.
	mux.HandleFunc("GET /register", h.redirectIfAuthed(h.RegisterForm))
	mux.HandleFunc("POST /register", h.redirectIfAuthed(h.Register))
	mux.HandleFunc("GET /login", h.redirectIfAuthed(h.LoginForm))
	mux.HandleFunc("POST /login", h.redirectIfAuthed(h.Login))
	mux.HandleFunc("POST /logout", h.Logout)

	// Защищённые страницы — оборачиваем в requireAuth.
	mux.HandleFunc("GET /expenses", h.requireAuth(h.ListExpenses))
	mux.HandleFunc("GET /expenses/new", h.requireAuth(h.NewExpenseForm))
	mux.HandleFunc("POST /expenses", h.requireAuth(h.CreateExpense))
	mux.HandleFunc("GET /expenses/{id}/edit", h.requireAuth(h.EditExpenseForm))
	mux.HandleFunc("POST /expenses/{id}", h.requireAuth(h.UpdateExpense))
	mux.HandleFunc("POST /expenses/{id}/delete", h.requireAuth(h.DeleteExpense))

	mux.HandleFunc("GET /categories", h.requireAuth(h.ListCategories))

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

// runMigrations выполняет все *.sql-файлы из папки dir в алфавитном порядке.
func runMigrations(db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(data)); err != nil {
			return err
		}
		log.Printf("миграция применена: %s", name)
	}
	return nil
}