package handler

import (
	"net/http"
	"time"

	"github.com/clube-do-livro/app/internal/middleware"
	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type meetingHandler struct{ d Deps }

func (h *meetingHandler) list(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	meets, err := h.d.Meetings.Upcoming(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	render(w, "encontros_lista", templ.PageData{
		Title: "Encontros", Active: "encontros", Me: me,
		Data: map[string]any{"Meetings": meets},
	})
}

func (h *meetingHandler) newPage(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	books, _ := h.d.Books.List(r.Context(), "")
	render(w, "encontros_novo", templ.PageData{
		Title: "Novo encontro", Active: "encontros", Me: me,
		Data: map[string]any{"Books": books},
	})
}

func (h *meetingHandler) create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// input type="datetime-local" envia "2006-01-02T15:04"
	dt, err := time.ParseInLocation("2006-01-02T15:04", r.FormValue("datetime"), time.Local)
	if err != nil {
		http.Error(w, "data inválida", 400)
		return
	}
	m := &model.Meeting{
		Title:    r.FormValue("title"),
		Datetime: dt,
		Location: r.FormValue("location"),
	}
	if b := r.FormValue("book_id"); b != "" {
		if id, err := uuid.Parse(b); err == nil {
			m.BookID = &id
		}
	}
	if err := h.d.Meetings.Create(r.Context(), m); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/encontros", http.StatusSeeOther)
}

func (h *meetingHandler) detail(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	meet, err := h.d.Meetings.Get(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	atts, _ := h.d.Meetings.Attendances(r.Context(), id)
	agenda, _ := h.d.Meetings.Agenda(r.Context(), id)

	var book *model.Book
	if meet.BookID != nil {
		book, _ = h.d.Books.Get(r.Context(), *meet.BookID)
	}
	totals := struct{ Confirmado, Talvez, NaoVou int }{}
	var minha string
	for _, a := range atts {
		switch a.Status {
		case model.PresencaConfirmado:
			totals.Confirmado++
		case model.PresencaTalvez:
			totals.Talvez++
		case model.PresencaNaoVou:
			totals.NaoVou++
		}
		if a.MemberID == me.ID {
			minha = a.Status
		}
	}

	render(w, "encontros_detalhe", templ.PageData{
		Title: meet.Title, Active: "encontros", Me: me,
		Data: map[string]any{
			"Meeting":       meet,
			"Book":          book,
			"Attendances":   atts,
			"Agenda":        agenda,
			"Totais":        totals,
			"MinhaPresenca": minha,
		},
	})
}

func (h *meetingHandler) setAttendance(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := h.d.Meetings.SetAttendance(r.Context(), id, me.ID, r.FormValue("status")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/encontros/"+id.String(), http.StatusSeeOther)
}

func (h *meetingHandler) addAgenda(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := h.d.Meetings.AddAgendaItem(r.Context(), id, me.ID, r.FormValue("content")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/encontros/"+id.String(), http.StatusSeeOther)
}
