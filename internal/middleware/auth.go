// Package middleware contém middlewares HTTP reutilizáveis.
package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/service"
)

// CookieName é o nome do cookie de sessão.
const CookieName = "clube_sess"

// contextKey evita colisões no context.
type contextKey string

const memberKey contextKey = "member"

// RequireAuth garante que o request tem um cookie de sessão válido.
// Redireciona para /login caso contrário.
func RequireAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(CookieName)
			if err != nil || c.Value == "" {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			m, err := auth.MemberFromToken(r.Context(), c.Value)
			if err != nil {
				if errors.Is(err, service.ErrNaoAutorizado) {
					http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", Path: "/", MaxAge: -1})
					http.Redirect(w, r, "/login", http.StatusFound)
					return
				}
				http.Error(w, "erro interno", 500)
				return
			}
			ctx := context.WithValue(r.Context(), memberKey, m)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin exige que o membro autenticado seja admin.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := MemberFrom(r.Context())
		if m == nil || !m.IsAdmin {
			http.Error(w, "acesso restrito a administradores", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// MemberFrom recupera o membro autenticado do contexto (nil se não autenticado).
func MemberFrom(ctx context.Context) *model.Member {
	m, _ := ctx.Value(memberKey).(*model.Member)
	return m
}
