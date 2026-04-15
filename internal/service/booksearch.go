package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/clube-do-livro/app/internal/model"
)

// BookSearch busca metadados de livros em fontes externas.
// Estratégia: tenta Google Books primeiro; se vier vazio OU falhar,
// cai para Open Library. Sempre retorna no máximo um resultado por fonte
// para simplificar a UI.
type BookSearch struct {
	HTTPClient      *http.Client
	GoogleURL       string // permite mock em testes
	OpenLibraryURL  string
	GoogleAPIKey    string
}

// NewBookSearch cria um BookSearch com endpoints reais.
func NewBookSearch(apiKey string) *BookSearch {
	return &BookSearch{
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		GoogleURL:      "https://www.googleapis.com/books/v1/volumes",
		OpenLibraryURL: "https://openlibrary.org/search.json",
		GoogleAPIKey:   apiKey,
	}
}

// Search devolve até N sugestões, tentando primeiro Google Books e, se
// não retornar nada útil, Open Library.
func (s *BookSearch) Search(ctx context.Context, query string) ([]model.ExternalBook, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query vazia")
	}
	res, err := s.searchGoogle(ctx, query)
	if err == nil && len(res) > 0 {
		return res, nil
	}
	// Fallback
	res2, err2 := s.searchOpenLibrary(ctx, query)
	if err2 != nil {
		if err != nil {
			return nil, fmt.Errorf("google: %v; open library: %w", err, err2)
		}
		return nil, err2
	}
	return res2, nil
}

// --- Google Books ---

type googleResp struct {
	Items []struct {
		VolumeInfo struct {
			Title         string   `json:"title"`
			Authors       []string `json:"authors"`
			Publisher     string   `json:"publisher"`
			PublishedDate string   `json:"publishedDate"`
			Description   string   `json:"description"`
			PageCount     int      `json:"pageCount"`
			ImageLinks    struct {
				Thumbnail string `json:"thumbnail"`
			} `json:"imageLinks"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

func (s *BookSearch) searchGoogle(ctx context.Context, query string) ([]model.ExternalBook, error) {
	u, _ := url.Parse(s.GoogleURL)
	q := u.Query()
	q.Set("q", query)
	q.Set("maxResults", "5")
	if s.GoogleAPIKey != "" {
		q.Set("key", s.GoogleAPIKey)
	}
	u.RawQuery = q.Encode()

	body, err := s.getJSON(ctx, u.String())
	if err != nil {
		return nil, err
	}
	var gr googleResp
	if err := json.Unmarshal(body, &gr); err != nil {
		return nil, err
	}
	var out []model.ExternalBook
	for _, it := range gr.Items {
		v := it.VolumeInfo
		year, _ := strconv.Atoi(firstYear(v.PublishedDate))
		out = append(out, model.ExternalBook{
			Title:     v.Title,
			Author:    strings.Join(v.Authors, ", "),
			CoverURL:  httpsify(v.ImageLinks.Thumbnail),
			Synopsis:  v.Description,
			Publisher: v.Publisher,
			Year:      year,
			Pages:     v.PageCount,
			Source:    "google_books",
		})
	}
	return out, nil
}

// --- Open Library ---

type openLibraryResp struct {
	Docs []struct {
		Title           string   `json:"title"`
		AuthorName      []string `json:"author_name"`
		Publisher       []string `json:"publisher"`
		FirstPublish    int      `json:"first_publish_year"`
		NumberOfPages   int      `json:"number_of_pages_median"`
		CoverI          int      `json:"cover_i"`
		FirstSentence   []string `json:"first_sentence"`
	} `json:"docs"`
}

func (s *BookSearch) searchOpenLibrary(ctx context.Context, query string) ([]model.ExternalBook, error) {
	u, _ := url.Parse(s.OpenLibraryURL)
	q := u.Query()
	q.Set("q", query)
	q.Set("limit", "5")
	u.RawQuery = q.Encode()

	body, err := s.getJSON(ctx, u.String())
	if err != nil {
		return nil, err
	}
	var or openLibraryResp
	if err := json.Unmarshal(body, &or); err != nil {
		return nil, err
	}
	var out []model.ExternalBook
	for _, d := range or.Docs {
		cover := ""
		if d.CoverI > 0 {
			cover = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", d.CoverI)
		}
		synopsis := ""
		if len(d.FirstSentence) > 0 {
			synopsis = d.FirstSentence[0]
		}
		publisher := ""
		if len(d.Publisher) > 0 {
			publisher = d.Publisher[0]
		}
		out = append(out, model.ExternalBook{
			Title:     d.Title,
			Author:    strings.Join(d.AuthorName, ", "),
			CoverURL:  cover,
			Synopsis:  synopsis,
			Publisher: publisher,
			Year:      d.FirstPublish,
			Pages:     d.NumberOfPages,
			Source:    "open_library",
		})
	}
	return out, nil
}

// --- helpers ---

func (s *BookSearch) getJSON(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// firstYear extrai o ano de strings como "1999-01-02" ou "1999".
func firstYear(s string) string {
	if len(s) >= 4 {
		return s[:4]
	}
	return "0"
}

// httpsify normaliza links http da API do Google para https, que
// pode ser mixed-content bloqueado no navegador.
func httpsify(u string) string {
	if strings.HasPrefix(u, "http://") {
		return "https://" + strings.TrimPrefix(u, "http://")
	}
	return u
}
