package model

import (
	"time"

	"github.com/google/uuid"
)

// Status de presença.
const (
	PresencaConfirmado = "confirmado"
	PresencaNaoVou    = "nao_vou"
	PresencaTalvez    = "talvez"
)

// Meeting é um encontro agendado pelo admin.
type Meeting struct {
	ID         uuid.UUID
	Title      string
	Datetime   time.Time
	Location   string
	BookID     *uuid.UUID
	RemindedAt *time.Time
	CreatedAt  time.Time
}

// Attendance é a resposta de um membro à presença no encontro.
type Attendance struct {
	ID        uuid.UUID
	MeetingID uuid.UUID
	MemberID  uuid.UUID
	Status    string
	// Name é preenchido em listagens para exibir o nome do membro.
	Name string
}

// AgendaItem é um tópico sugerido por um membro para o encontro.
type AgendaItem struct {
	ID        uuid.UUID
	MeetingID uuid.UUID
	MemberID  uuid.UUID
	Content   string
	CreatedAt time.Time
	// Author preenchido em listagens.
	Author string
}
