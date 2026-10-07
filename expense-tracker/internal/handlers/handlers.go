package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"expense-tracker/internal/repository"
)

type Handler struct {
	expenses   *repository.ExpenseRepository
	categories *repository.CategoryRepository
	users      *repository.UserRepository
	sessions   *repository.SessionRepository
	templates  map[string]*template.Template
}

func New(
	er *repository.ExpenseRepository,
	cr *repository.CategoryRepository,
	ur *repository.UserRepository,
	sr *repository.SessionRepository,
) (*Handler, error) {
	h := &Handler{expenses: er, categories: cr, users: ur, sessions: sr}
	if err := h.parseTemplates("web/templates"); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Handler) parseTemplates(dir string) error {
	pages := []string{
		"index", "expenses", "expense_form", "categories",
		"register", "login",
	}
	h.templates = make(map[string]*template.Template)
	for _, p := range pages {
		t, err := template.ParseFiles(
			filepath.Join(dir, "layout.html"),
			filepath.Join(dir, p+".html"),
		)
		if err != nil {
			return err
		}
		h.templates[p] = t
	}
	return nil
}

func (h *Handler) render(w http.ResponseWriter, page string, data any) {
	h.renderStatus(w, page, data, http.StatusOK)
}

func (h *Handler) renderStatus(w http.ResponseWriter, page string, data any, status int) {
	t, ok := h.templates[page]
	if !ok {
		http.Error(w, "шаблон не найден: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ошибка рендера шаблона %s: %v", page, err)
	}
}

// Home — публичная главная. Если пользователь залогинен,
// покажем это в шапке (см. layout).
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	h.render(w, "index", nil)
}