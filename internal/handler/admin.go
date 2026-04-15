package handler

import (
	"net/http"

	"github.com/clube-do-livro/app/internal/middleware"
	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/templ"
	"github.com/google/uuid"
)

type adminHandler struct{ d Deps }

func (h *adminHandler) home(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	members, _ := h.d.Auth.Members.ListMembers(r.Context())
	sugeridos, _ := h.d.Books.List(r.Context(), model.StatusSugerido)
	lendo, _ := h.d.Books.List(r.Context(), model.StatusLendoAgora)
	round, candidatos, _ := h.d.Voting.CurrentRound(r.Context())

	render(w, "admin", templ.PageData{
		Title: "Admin", Active: "admin", Me: me,
		Data: map[string]any{
			"Members":    members,
			"Sugeridos":  sugeridos,
			"LendoAgora": lendo,
			"Aberta":     round != nil,
			"Candidatos": candidatos,
		},
	})
}

func (h *adminHandler) createMember(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	_, err := h.d.Auth.CreateMember(r.Context(),
		r.FormValue("name"), r.FormValue("email"), r.FormValue("password"),
		r.FormValue("is_admin") != "")
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *adminHandler) openRound(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	var ids []uuid.UUID
	for _, v := range r.Form["book_ids"] {
		if id, err := uuid.Parse(v); err == nil {
			ids = append(ids, id)
		}
	}
	if _, err := h.d.Voting.OpenRound(r.Context(), ids); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *adminHandler) closeRound(w http.ResponseWriter, r *http.Request) {
	round, _, err := h.d.Voting.CurrentRound(r.Context())
	if err != nil || round == nil {
		http.Error(w, "não há rodada aberta", 400)
		return
	}
	if _, err := h.d.Voting.Close(r.Context(), round.ID); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *adminHandler) markLido(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	id, err := uuid.Parse(r.FormValue("book_id"))
	if err != nil {
		http.Error(w, "book_id inválido", 400)
		return
	}
	if err := h.d.Books.MarkLido(r.Context(), id); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
