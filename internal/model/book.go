package model

import (
	"time"

	"github.com/google/uuid"
)

// Status possíveis para um livro.
const (
	StatusSugerido   = "sugerido"
	StatusEmVotacao  = "em_votacao"
	StatusLendoAgora = "lendo_agora"
	StatusLido       = "lido"
)

// Book é um livro no catálogo do clube (sugestão, em leitura ou lido).
type Book struct {
	ID          uuid.UUID
	Title       string
	Author      string
	CoverURL    string
	Synopsis    string
	Publisher   string
	Year        int
	Pages       int
	Status      string
	SuggestedBy     *uuid.UUID
	SuggestedByName string
	FinishedAt      *time.Time
	CreatedAt   time.Time
	Tags        []string
}

// ExternalBook é o resultado vindo das APIs externas (Google Books / Open Library),
// antes de virar um Book persistido.
type ExternalBook struct {
	Title     string
	Author    string
	CoverURL  string
	Synopsis  string
	Publisher string
	Year      int
	Pages     int
	Source    string // "google_books" ou "open_library"
}
