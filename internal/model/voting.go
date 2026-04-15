package model

import (
	"time"

	"github.com/google/uuid"
)

// Status de rodada.
const (
	RoundAberta    = "aberta"
	RoundEncerrada = "encerrada"
)

// VoteRound é uma rodada de votação.
type VoteRound struct {
	ID           uuid.UUID
	Status       string
	OpenedAt     time.Time
	ClosedAt     *time.Time
	WinnerBookID *uuid.UUID
	Books        []Book // livros candidatos (opcionalmente preenchido)
}

// Vote é um voto único de um membro numa rodada.
type Vote struct {
	ID        uuid.UUID
	RoundID   uuid.UUID
	MemberID  uuid.UUID
	BookID    uuid.UUID
	CreatedAt time.Time
}

// VoteCount agrega a contagem de votos de um livro em uma rodada.
type VoteCount struct {
	BookID uuid.UUID
	Title  string
	Count  int
}
