package model

import (
	"time"

	"github.com/google/uuid"
)

// Review é uma avaliação de um membro sobre um livro lido.
// Todos os campos numéricos são ponteiros porque são opcionais.
type Review struct {
	ID              uuid.UUID
	BookID          uuid.UUID
	MemberID        uuid.UUID
	NotaGeral       *int
	NotaEscrita     *int
	NotaEnredo      *int
	NotaExpectativa *int
	ReviewText      string
	HasSpoiler      bool
	Citacao         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ReviewStats é a média agregada de um livro, considerando apenas
// notas preenchidas (NULLs são ignorados).
type ReviewStats struct {
	BookID          uuid.UUID
	AvgGeral        float64
	AvgEscrita      float64
	AvgEnredo       float64
	AvgExpectativa  float64
	CountGeral      int
	CountEscrita    int
	CountEnredo     int
	CountExpectativa int
	TotalReviews    int
}
