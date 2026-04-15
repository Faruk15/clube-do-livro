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

func (h *authHandler) showSignup(w http.ResponseWriter, r *http.Request) {
	render(w, "signup", templ.PageData{Title: "Criar conta"})
}

func (h *authHandler) doSignup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	name := r.FormValue("name")
	email := r.FormValue("email")
	pass := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	fail := func(msg string) {
		render(w, "signup", templ.PageData{
			Title: "Criar conta", Flash: msg,
			Data: map[string]any{"Name": name, "Email": email},
		})
	}

	if pass != confirm {
		fail("As senhas não coincidem.")
		return
	}
	if len(pass) < 6 {
		fail("A senha deve ter ao menos 6 caracteres.")
		return
	}

	if _, err := h.d.Auth.CreateMember(r.Context(), name, email, pass, false); err != nil {
		switch {
		case errors.Is(err, service.ErrEmailJaExiste):
			fail("Já existe uma conta com este e-mail.")
		default:
			fail("Não foi possível criar a conta: " + err.Error())
		}
		return
	}

	// Auto-login após o cadastro.
	token, _, err := h.d.Auth.Login(r.Context(), email, pass)
	if err != nil {
		// Conta criada, mas o login automático falhou — manda para o login manual.
		http.Redirect(w, r, "/login", http.StatusFound)
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
