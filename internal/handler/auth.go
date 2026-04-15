package handler

import (
	"errors"
	"net/http"

	"github.com/clube-do-livro/app/internal/middleware"
	"github.com/clube-do-livro/app/internal/service"
	"github.com/clube-do-livro/app/internal/templ"
)

type authHandler struct{ d Deps }

func (h *authHandler) showLogin(w http.ResponseWriter, r *http.Request) {
	render(w, "login", templ.PageData{Title: "Entrar"})
}

func (h *authHandler) doLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	email := r.FormValue("email")
	pass := r.FormValue("password")
	token, _, err := h.d.Auth.Login(r.Context(), email, pass)
	if err != nil {
		flash := "Credenciais inválidas."
		if !errors.Is(err, service.ErrCredenciaisInvalidas) {
			flash = "Erro ao fazer login."
		}
		render(w, "login", templ.PageData{Title: "Entrar", Flash: flash})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *authHandler) doLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(middleware.CookieName); err == nil {
		_ = h.d.Auth.Logout(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: middleware.CookieName, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusFound)
}
