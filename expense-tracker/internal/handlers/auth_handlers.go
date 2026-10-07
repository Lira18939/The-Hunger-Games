package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"expense-tracker/internal/repository"
)

const (
	sessionCookieName = "session_id"
	sessionTTL        = 7 * 24 * time.Hour // неделя
)

type AuthPageData struct {
	Email string
	Error string
	IsRegister bool
}

// GET /register
func (h *Handler) RegisterForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "register", AuthPageData{IsRegister: true})
}

// POST /register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "ошибка разбора формы", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	renderErr := func(msg string) {
		h.renderStatus(w, "register", AuthPageData{
			Email: email, Error: msg, IsRegister: true,
		}, http.StatusBadRequest)
	}

	if email == "" || !strings.Contains(email, "@") {
		renderErr("Введите корректный email")
		return
	}
	if len(password) < 6 {
		renderErr("Пароль должен быть не короче 6 символов")
		return
	}
	if password != password2 {
		renderErr("Пароли не совпадают")
		return
	}

	if _, err := h.users.GetByEmail(email); err == nil {
		renderErr("Пользователь с таким email уже зарегистрирован")
		return
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		renderErr("Ошибка БД: " + err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		renderErr("Не удалось захешировать пароль")
		return
	}
	userID, err := h.users.Create(email, string(hash))
	if err != nil {
		renderErr("Не удалось создать пользователя: " + err.Error())
		return
	}

	if err := h.startSession(w, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// GET /login
func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "login", AuthPageData{})
}

// POST /login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "ошибка разбора формы", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	renderErr := func(msg string) {
		h.renderStatus(w, "login", AuthPageData{
			Email: email, Error: msg,
		}, http.StatusUnauthorized)
	}

	u, err := h.users.GetByEmail(email)
	if err != nil {
		// Намеренно не говорим, что именно неверно: email или пароль.
		renderErr("Неверный email или пароль")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		renderErr("Неверный email или пароль")
		return
	}

	if err := h.startSession(w, u.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// POST /logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookieName)
	if err == nil {
		_ = h.sessions.Delete(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// startSession — вспомогательная функция: создаёт сессию и ставит cookie.
func (h *Handler) startSession(w http.ResponseWriter, userID int64) error {
	sess, err := h.sessions.Create(userID, sessionTTL)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sess.SessionID,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}