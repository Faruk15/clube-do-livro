package model

import (
	"time"

	"github.com/google/uuid"
)

// Member representa um usuário autenticado do clube.
type Member struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    time.Time
}

// Session é a sessão persistida (o cookie carrega o Token).
type Session struct {
	ID        uuid.UUID
	MemberID  uuid.UUID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
