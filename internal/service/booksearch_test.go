package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const googleSample = `{
  "items": [{
    "volumeInfo": {
      "title": "Dom Casmurro",
      "authors": ["Machado de Assis"],
      "publisher": "Companhia",
      "publishedDate": "1899-01-01",
      "description": "Clássico nacional.",
      "pageCount": 256,
      "imageLinks": {"thumbnail": "http://books.google.com/cover.png"}
    }
  }]
}`

const googleEmpty = `{"items": []}`

const openLibrarySample = `{
  "docs": [{
    "title": "Dom Casmurro",
    "author_name": ["Machado de Assis"],
    "publisher": ["Companhia"],
    "first_publish_year": 1899,
    "number_of_pages_median": 256,
    "cover_i": 1234,
    "first_sentence": ["Uma noite destas..."]
  }]
}`

func TestBookSearch_GoogleSucesso(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.URL.RawQuery, "q=dom+casmurro")
		_, _ = w.Write([]byte(googleSample))
	}))
	defer srv.Close()

	s := &BookSearch{HTTPClient: srv.Client(), GoogleURL: srv.URL, OpenLibraryURL: "http://ignored"}
	res, err := s.Search(context.Background(), "dom casmurro")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "Dom Casmurro", res[0].Title)
	require.Equal(t, "Machado de Assis", res[0].Author)
	require.Equal(t, 1899, res[0].Year)
	require.Equal(t, 256, res[0].Pages)
	// http:// vira https://
	require.Equal(t, "https://books.google.com/cover.png", res[0].CoverURL)
	require.Equal(t, "google_books", res[0].Source)
}

func TestBookSearch_FallbackParaOpenLibrary(t *testing.T) {
	// Google retorna vazio -> deve cair para Open Library
	google := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(googleEmpty))
	}))
	defer google.Close()

	openlib := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(openLibrarySample))
	}))
	defer openlib.Close()

	s := &BookSearch{HTTPClient: google.Client(), GoogleURL: google.URL, OpenLibraryURL: openlib.URL}
	res, err := s.Search(context.Background(), "dom casmurro")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "Dom Casmurro", res[0].Title)
	require.Equal(t, "open_library", res[0].Source)
	require.Contains(t, res[0].CoverURL, "covers.openlibrary.org")
}

func TestBookSearch_FallbackQuandoGoogleErra(t *testing.T) {
	// Google retorna 500 -> deve tentar Open Library
	google := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer google.Close()

	openlib := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(openLibrarySample))
	}))
	defer openlib.Close()

	s := &BookSearch{HTTPClient: google.Client(), GoogleURL: google.URL, OpenLibraryURL: openlib.URL}
	res, err := s.Search(context.Background(), "q")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "open_library", res[0].Source)
}

func TestBookSearch_AmbosFalham(t *testing.T) {
	fail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer fail.Close()

	s := &BookSearch{HTTPClient: fail.Client(), GoogleURL: fail.URL, OpenLibraryURL: fail.URL}
	_, err := s.Search(context.Background(), "q")
	require.Error(t, err)
}

func TestBookSearch_QueryVazia(t *testing.T) {
	s := &BookSearch{HTTPClient: http.DefaultClient}
	_, err := s.Search(context.Background(), "   ")
	require.Error(t, err)
}
