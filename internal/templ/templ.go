// Package templ embute e renderiza os templates HTML do clube.
//
// Nota arquitetural: o layout segue a filosofia Templ (SSR tipado + HTMX),
// mas o rendering é feito com html/template da stdlib, mantendo o
// projeto sem dependências de geração de código externas.
package templ

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
)

//go:embed *.gohtml
var files embed.FS

// Funcs expostas aos templates.
var funcMap = template.FuncMap{
	"fmtDate": func(t time.Time) string { return t.Format("02/01/2006") },
	"fmtDateTime": func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("02/01/2006 15:04")
	},
	"fmtDateTimeInput": func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02T15:04")
	},
	"fmtAvg": func(v float64) string { return fmt.Sprintf("%.1f", v) },
	"stars": func(n int) string {
		s := ""
		for i := 1; i <= 5; i++ {
			if i <= n {
				s += "★"
			} else {
				s += "☆"
			}
		}
		return s
	},
	"star": func(v float64) string {
		n := int(v + 0.5)
		s := ""
		for i := 1; i <= 5; i++ {
			if i <= n {
				s += "★"
			} else {
				s += "☆"
			}
		}
		return s
	},
	"ifnil": func(p *int) int {
		if p == nil {
			return 0
		}
		return *p
	},
	"statusLabel": func(s string) string {
		switch s {
		case model.StatusSugerido:
			return "sugerido"
		case model.StatusEmVotacao:
			return "em votação"
		case model.StatusLendoAgora:
			return "lendo agora"
		case model.StatusLido:
			return "lido"
		}
		return s
	},
	"presencaLabel": func(s string) string {
		switch s {
		case model.PresencaConfirmado:
			return "confirmado"
		case model.PresencaNaoVou:
			return "não vou"
		case model.PresencaTalvez:
			return "talvez"
		}
		return s
	},
	"uuidStr": func(id uuid.UUID) string { return id.String() },
	"uuidPtr": func(p *uuid.UUID) string {
		if p == nil {
			return ""
		}
		return p.String()
	},
	"dict": func(vals ...any) map[string]any {
		m := map[string]any{}
		for i := 0; i+1 < len(vals); i += 2 {
			key, _ := vals[i].(string)
			m[key] = vals[i+1]
		}
		return m
	},
	"hasBook": func(bs []model.Book, id uuid.UUID) bool {
		for _, b := range bs {
			if b.ID == id {
				return true
			}
		}
		return false
	},
}

// Cada página tem sua própria árvore de templates clonada do base.
// Isso evita que o {{define "content"}} de um arquivo sobrescreva o de outro
// (no html/template, o último parseado vence quando há nomes duplicados).
var pages map[string]*template.Template

// Arquivos compartilhados (layout + parciais/fragmentos sem "content" próprio).
var sharedFiles = []string{
	"layout.gohtml",
	"bookcard.gohtml",
	"livros_buscar.gohtml",
}

// base contém layout + parciais/fragmentos compartilhados (sem "content" próprio).
var base *template.Template

func init() {
	base = template.Must(template.New("base").Funcs(funcMap).ParseFS(files, sharedFiles...))

	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		panic("falha ao listar templates: " + err.Error())
	}

	shared := map[string]bool{}
	for _, s := range sharedFiles {
		shared[s] = true
	}

	// Nomes de template já definidos no base (para identificar os novos por arquivo).
	baseNames := map[string]bool{}
	for _, t := range base.Templates() {
		baseNames[t.Name()] = true
	}

	pages = map[string]*template.Template{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".gohtml") || shared[name] {
			continue
		}
		clone := template.Must(base.Clone())
		clone = template.Must(clone.ParseFS(files, name))
		// Indexa cada template novo desse arquivo (ex: "encontros_lista", "content").
		for _, t := range clone.Templates() {
			n := t.Name()
			if baseNames[n] || n == "content" {
				continue
			}
			pages[n] = clone
		}
	}
}

// Render escreve a view pedida usando o layout base.
// Para HTMX swaps, use Render com o nome do fragmento (ex: "fragment_buscar")
// — fragmentos vivem no base compartilhado.
func Render(w io.Writer, name string, data any) error {
	if p, ok := pages[name]; ok {
		return p.ExecuteTemplate(w, name, data)
	}
	// Fallback: fragmentos / parciais definidos nos arquivos compartilhados.
	if base.Lookup(name) != nil {
		return base.ExecuteTemplate(w, name, data)
	}
	return fmt.Errorf("template %q desconhecido", name)
}

// PageData é o envelope passado para as views de página completa.
type PageData struct {
	Title  string
	Active string // id da aba ativa no menu
	Me     *model.Member
	Flash  string // mensagem simples (erro ou sucesso)
	Data   any
}
