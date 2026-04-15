package handler

import (
	"net/http"

	"github.com/clube-do-livro/app/internal/middleware"
	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/templ"
)

type dashboardHandler struct{ d Deps }

func (h *dashboardHandler) show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	me := middleware.MemberFrom(ctx)

	lendo, _ := h.d.Books.List(ctx, model.StatusLendoAgora)
	rodada, roLivros, _ := h.d.Voting.CurrentRound(ctx)
	upcoming, _ := h.d.Meetings.Upcoming(ctx)
	var proximo *model.Meeting
	if len(upcoming) > 0 {
		proximo = &upcoming[0]
	}

	render(w, "dashboard", templ.PageData{
		Title: "Dashboard", Active: "dashboard", Me: me,
		Data: map[string]any{
			"Lendo":         lendo,
			"RodadaAberta":  rodada != nil,
			"RodadaLivros":  roLivros,
			"Proximo":       proximo,
		},
	})
}
