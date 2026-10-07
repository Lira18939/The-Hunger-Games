package handlers

import (
	"context"
	"net/http"

	"expense-tracker/internal/repository"
)

type contextKey string

const userIDKey contextKey = "userID"

// userIDFromContext — читаем ID пользователя из контекста запроса.
func userIDFromContext(r *http.Request) (int64, bool) {
	v := r.Context().Value(userIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// requireAuth — middleware для защиты страниц.
// Если пользователь не авторизован — редирект на /login.
func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		sess, err := h.sessions.GetValid(c.Value)
		if err != nil {
			// Сессия истекла или не найдена — удаляем cookie и на логин.
			http.SetCookie(w, &http.Cookie{
				Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1,
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Кладём user_id в контекст — дальше его достанут обработчики.
		ctx := context.WithValue(r.Context(), userIDKey, sess.UserID)
		next(w, r.WithContext(ctx))
	}
}

// redirectIfAuthed — для /login и /register: если уже вошли, идём на /expenses.
func (h *Handler) redirectIfAuthed(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookieName)
		if err == nil {
			if _, err := h.sessions.GetValid(c.Value); err == nil {
				http.Redirect(w, r, "/expenses", http.StatusSeeOther)
				return
			}
		}
		next(w, r)
	}
}

// Оставим ссылку на пакет repository, чтобы goimports не удалил импорт,
// если он не используется напрямую (на будущее для этапа 7).
var _ = repository.ErrSessionNotFound