// Package handler reúne os handlers HTTP e monta o roteador.
package handler

import (
	"net/http"

	"github.com/clube-do-livro/app/internal/middleware"
	"github.com/clube-do-livro/app/internal/service"
	"github.com/clube-do-livro/app/internal/templ"
	"github.com/go-chi/chi/v5"
)

// Deps é o conjunto de dependências passadas para os handlers.
type Deps struct {
	Auth       *service.AuthService
	Books      *service.BookService
	Search     *service.BookSearch
	Voting     *service.VotingService
	Reviews    *service.ReviewService
	Meetings   *service.MeetingService
}

// New monta o roteador completo da aplicação.
func New(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Get("/login", (&authHandler{d}).showLogin)
	r.Post("/login", (&authHandler{d}).doLogin)
	r.Get("/logout", (&authHandler{d}).doLogout)
	r.Get("/signup", (&authHandler{d}).showSignup)
	r.Post("/signup", (&authHandler{d}).doSignup)

	// Área autenticada.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(d.Auth))

		r.Get("/", (&dashboardHandler{d}).show)

		bh := &bookHandler{d}
		r.Get("/livros", bh.list)
		r.Get("/livros/sugerir", bh.suggestPage)
		r.Post("/livros/buscar", bh.search)
		r.Post("/livros/sugerir", bh.submitSuggest)
		r.Get("/livros/{id}", bh.detail)
		r.Post("/livros/{id}/remover-sugestao", bh.removeSuggestion)
		r.Post("/livros/{id}/tag", bh.addTag)

		vh := &votingHandler{d}
		r.Get("/votacao", vh.show)
		r.Post("/votacao/votar", vh.vote)

		rh := &reviewHandler{d}
		r.Get("/avaliacoes", rh.list)
		r.Post("/avaliacoes", rh.submit)

		mh := &meetingHandler{d}
		r.Get("/encontros", mh.list)
		r.Get("/encontros/{id}", mh.detail)
		r.Post("/encontros/{id}/presenca", mh.setAttendance)
		r.Post("/encontros/{id}/pauta", mh.addAgenda)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Get("/encontros/novo", mh.newPage)
			r.Post("/encontros", mh.create)
		})

		r.Get("/historico", bh.historico)

		// Admin.
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			ah := &adminHandler{d}
			r.Get("/", ah.home)
			r.Post("/membros", ah.createMember)
			r.Post("/votacao/abrir", ah.openRound)
			r.Post("/votacao/encerrar", ah.closeRound)
			r.Post("/livros/concluir", ah.markLido)
		})
	})

	return r
}

// render é um atalho para renderizar um template com PageData.
func render(w http.ResponseWriter, name string, pd templ.PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templ.Render(w, name, pd); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
