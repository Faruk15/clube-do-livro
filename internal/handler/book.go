package handler

import (
	"net/http"
	"strconv"

	"github.com/clube-do-livro/app/internal/middleware"
	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type bookHandler struct{ d Deps }

func (h *bookHandler) list(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	books, err := h.d.Books.List(r.Context(), r.URL.Query().Get("status"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	render(w, "livros_list", templ.PageData{
		Title: "Livros", Active: "livros", Me: me,
		Data: map[string]any{"Books": books},
	})
}

func (h *bookHandler) suggestPage(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	render(w, "livros_sugerir", templ.PageData{
		Title: "Sugerir livro", Active: "livros", Me: me,
	})
}

// search: chamada via HTMX; devolve só o fragmento HTML.
func (h *bookHandler) search(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	q := r.FormValue("q")
	results, err := h.d.Search.Search(r.Context(), q)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<p style="color:#c0392b">Erro ao buscar: ` + err.Error() + `</p>`))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templ.Render(w, "fragment_buscar", map[string]any{"Results": results})
}

func (h *bookHandler) submitSuggest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	me := middleware.MemberFrom(r.Context())
	year, _ := strconv.Atoi(r.FormValue("year"))
	pages, _ := strconv.Atoi(r.FormValue("pages"))
	b := &model.Book{
		Title:     r.FormValue("title"),
		Author:    r.FormValue("author"),
		CoverURL:  r.FormValue("cover_url"),
		Synopsis:  r.FormValue("synopsis"),
		Publisher: r.FormValue("publisher"),
		Year:      year,
		Pages:     pages,
	}
	if _, err := h.d.Books.Suggest(r.Context(), me, b); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/livros", http.StatusSeeOther)
}

func (h *bookHandler) detail(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	b, err := h.d.Books.Get(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	stats, reviews, _ := h.d.Reviews.Stats(r.Context(), id)
	minha, _ := h.d.Reviews.Mine(r.Context(), id, me.ID)
	// O template acessa Minha.* incondicionalmente (Go templates não fazem
	// short-circuit em `and`); um valor zero evita nil pointer quando o
	// usuário ainda não avaliou.
	if minha == nil {
		minha = &model.Review{}
	}

	render(w, "livros_detalhe", templ.PageData{
		Title: b.Title, Active: "livros", Me: me,
		Data: map[string]any{
			"Book":      b,
			"Stats":     stats,
			"Reviews":   reviews,
			"Minha":     minha,
			"CanReview": b.Status == model.StatusLido,
			"Ratings":   []int{1, 2, 3, 4, 5},
		},
	})
}

func (h *bookHandler) addTag(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := h.d.Books.AddTag(r.Context(), id, r.FormValue("tag")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/livros/"+id.String(), http.StatusSeeOther)
}

func (h *bookHandler) historico(w http.ResponseWriter, r *http.Request) {
	me := middleware.MemberFrom(r.Context())
	books, err := h.d.Books.ListFinished(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	render(w, "historico", templ.PageData{
		Title: "Histórico", Active: "historico", Me: me,
		Data: map[string]any{"Books": books},
	})
}
