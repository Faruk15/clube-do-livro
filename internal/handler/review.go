package handler

import (
	"net/http"
	"strconv"

	"github.com/clube-do-livro/app/internal/middleware"
	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/templ"
	"github.com/google/uuid"
)

type reviewHandler struct{ d Deps }

func (h *reviewHandler) list(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	lidos, err := h.d.Books.List(r.Context(), model.StatusLido)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	render(w, "avaliacoes", templ.PageData{
		Title: "Avaliações", Active: "avaliacoes", Me: me,
		Data: map[string]any{"Lidos": lidos},
	})
}

// submit grava (ou atualiza) a avaliação do membro.
func (h *reviewHandler) submit(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	bookID, err := uuid.Parse(r.FormValue("book_id"))
	if err != nil {
		http.Error(w, "book_id inválido", 400)
		return
	}
	review := &model.Review{
		BookID:          bookID,
		NotaGeral:       parseOptInt(r.FormValue("nota_geral")),
		NotaEscrita:     parseOptInt(r.FormValue("nota_escrita")),
		NotaEnredo:      parseOptInt(r.FormValue("nota_enredo")),
		NotaExpectativa: parseOptInt(r.FormValue("nota_expectativa")),
		ReviewText:      r.FormValue("review_text"),
		HasSpoiler:      r.FormValue("has_spoiler") != "",
		Citacao:         r.FormValue("citacao"),
	}
	if err := h.d.Reviews.Upsert(r.Context(), me.ID, review); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/livros/"+bookID.String(), http.StatusSeeOther)
}

// parseOptInt converte "" em nil (campo opcional) e texto em *int.
func parseOptInt(s string) *int {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}
