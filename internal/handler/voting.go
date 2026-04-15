package handler

import (
	"net/http"

	"github.com/clube-do-livro/app/internal/middleware"
	"github.com/clube-do-livro/app/internal/templ"
	"github.com/google/uuid"
)

type votingHandler struct{ d Deps }

func (h *votingHandler) show(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	round, books, err := h.d.Voting.CurrentRound(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := map[string]any{
		"Aberta": round != nil,
		"Round":  round,
		"Livros": books,
	}

	if round != nil {
		vt, _ := h.d.Voting.MyVote(r.Context(), round.ID, me.ID)
		if vt != nil {
			data["MinhaEscolha"] = &vt.BookID
			for _, b := range books {
				if b.ID == vt.BookID {
					data["MinhaEscolhaTitulo"] = b.Title
				}
			}
		}
	}

	// Mostra a apuração APENAS da última rodada encerrada (sem livros em aberto).
	if round == nil {
		// nenhuma apuração especial; mantemos simples
	}

	render(w, "votacao", templ.PageData{
		Title: "Votação", Active: "votacao", Me: me, Data: data,
	})
}

func (h *votingHandler) vote(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	roundID, err1 := uuid.Parse(r.FormValue("round_id"))
	bookID, err2 := uuid.Parse(r.FormValue("book_id"))
	if err1 != nil || err2 != nil {
		http.Error(w, "parâmetros inválidos", 400)
		return
	}
	if err := h.d.Voting.Cast(r.Context(), roundID, me.ID, bookID); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/votacao", http.StatusSeeOther)
}
