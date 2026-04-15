package templ

import (
	"bytes"
	"strings"
	"testing"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRenderVotacaoSemRodada(t *testing.T) {
	me := &model.Member{ID: uuid.New(), Name: "Ana", Email: "ana@x", IsAdmin: false}
	pd := PageData{
		Title:  "Votação",
		Active: "votacao",
		Me:     me,
		Data: map[string]any{
			"Aberta": false,
			"Livros": []model.Book{},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, "votacao", pd))
	out := buf.String()
	require.Contains(t, out, "Clube do Livro")
	require.Contains(t, out, "Ana", "o nome do membro deveria aparecer na sidebar")
	require.Contains(t, out, "Dashboard", "o menu de navegação deveria aparecer")
	require.Contains(t, out, "Nenhuma rodada")
}

func TestRenderVotacaoComRodadaAberta(t *testing.T) {
	// Regressão: o template usava $.MinhaEscolha (raiz = PageData) em vez de
	// $.Data.MinhaEscolha, quebrando a renderização sempre que havia rodada aberta.
	me := &model.Member{ID: uuid.New(), Name: "Bia", Email: "bia@x"}
	bookID := uuid.New()
	round := &model.VoteRound{ID: uuid.New()}
	pd := PageData{
		Title:  "Votação",
		Active: "votacao",
		Me:     me,
		Data: map[string]any{
			"Aberta":             true,
			"Round":              round,
			"Livros":             []model.Book{{ID: bookID, Title: "Dom Casmurro", Author: "Machado"}},
			"MinhaEscolha":       &bookID,
			"MinhaEscolhaTitulo": "Dom Casmurro",
		},
	}
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, "votacao", pd))
	out := buf.String()
	require.Contains(t, out, "Dom Casmurro")
	require.Contains(t, out, "Rodada aberta")
	require.Contains(t, out, "Seu voto atual")
}

func TestRenderLoginNaoVazaConteudoDeOutraPagina(t *testing.T) {
	// Regressão: antes, todos os arquivos definiam {{define "content"}} e o
	// último parseado vencia, fazendo /login renderizar o conteúdo de outra view.
	pd := PageData{Title: "Entrar", Active: "", Me: nil, Data: nil}
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, "login", pd))
	out := buf.String()
	require.Contains(t, out, `name="email"`, "form de login deveria estar presente")
	require.Contains(t, out, `name="password"`, "campo de senha deveria estar presente")
	require.NotContains(t, out, "Votação", "conteúdo de outra página não deveria vazar")
	require.NotContains(t, out, "Nenhuma rodada", "conteúdo de outra página não deveria vazar")
}

func TestRenderLivrosDetalheSemMinha(t *testing.T) {
	// Regressão: usuário sem review ainda — template não pode panicar
	// acessando ponteiros nil. Handler passa &model.Review{} como zero.
	me := &model.Member{ID: uuid.New(), Name: "Dani", Email: "dani@x"}
	pd := PageData{
		Title:  "Livro",
		Active: "livros",
		Me:     me,
		Data: map[string]any{
			"Book":      &model.Book{ID: uuid.New(), Title: "Solaris", Author: "Lem", Status: model.StatusLido},
			"Stats":     model.ReviewStats{},
			"Reviews":   []model.Review{},
			"Minha":     &model.Review{}, // zero, não nil
			"CanReview": true,
			"Ratings":   []int{1, 2, 3, 4, 5},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, "livros_detalhe", pd))
	require.Contains(t, buf.String(), "Solaris")
}

func TestRenderLivrosDetalheComMinha(t *testing.T) {
	// Regressão: o template usava $.Minha (raiz = PageData) em vez de
	// $.Data.Minha, quebrando a renderização da página de detalhe.
	me := &model.Member{ID: uuid.New(), Name: "Caio", Email: "caio@x"}
	nota := 4
	pd := PageData{
		Title:  "Livro X",
		Active: "livros",
		Me:     me,
		Data: map[string]any{
			"Book":      &model.Book{ID: uuid.New(), Title: "1984", Author: "Orwell", Status: model.StatusLido},
			"Stats":     model.ReviewStats{},
			"Reviews":   []model.Review{},
			"Minha":     &model.Review{NotaGeral: &nota, NotaEscrita: &nota},
			"CanReview": true,
			"Ratings":   []int{1, 2, 3, 4, 5},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, "livros_detalhe", pd))
	out := buf.String()
	require.Contains(t, out, "1984")
	require.Contains(t, out, "Minha avaliação")
}

func TestRenderDashboard(t *testing.T) {
	me := &model.Member{ID: uuid.New(), Name: "Zé", Email: "ze@x"}
	pd := PageData{Title: "x", Active: "dashboard", Me: me, Data: map[string]any{}}
	var buf bytes.Buffer
	err := Render(&buf, "dashboard", pd)
	if err != nil {
		t.Log("erro:", err)
	}
	require.NoError(t, err)
	if !strings.Contains(buf.String(), "Zé") {
		t.Log(buf.String()[:800])
		t.Fatal("esperado 'Zé' na saída")
	}
}
