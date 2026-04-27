// Package store define as interfaces de persistência do clube.
// Há uma implementação Postgres neste pacote. Nos testes, os serviços
// usam implementações fake (ver *_test.go).
package store

import (
	"context"
	"errors"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
)

// ErrNotFound é retornado quando a entidade não existe.
var ErrNotFound = errors.New("não encontrado")

// MemberStore — persistência de membros e sessões.
type MemberStore interface {
	CreateMember(ctx context.Context, m *model.Member) error
	GetMemberByEmail(ctx context.Context, email string) (*model.Member, error)
	GetMemberByID(ctx context.Context, id uuid.UUID) (*model.Member, error)
	ListMembers(ctx context.Context) ([]model.Member, error)

	CreateSession(ctx context.Context, s *model.Session) error
	GetSessionByToken(ctx context.Context, token string) (*model.Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context, now time.Time) error
}

// BookStore — persistência de livros e tags.
type BookStore interface {
	CreateBook(ctx context.Context, b *model.Book) error
	GetBook(ctx context.Context, id uuid.UUID) (*model.Book, error)
	UpdateBookStatus(ctx context.Context, id uuid.UUID, status string, finishedAt *time.Time) error
	ListBooks(ctx context.Context, statusFilter string) ([]model.Book, error)
	ListFinished(ctx context.Context) ([]model.Book, error)

	DeleteBook(ctx context.Context, id uuid.UUID) error

	AddTag(ctx context.Context, bookID uuid.UUID, tag string) error
	RemoveTag(ctx context.Context, bookID uuid.UUID, tag string) error
	ListTags(ctx context.Context, bookID uuid.UUID) ([]string, error)
}

// VotingStore — persistência de rodadas e votos.
type VotingStore interface {
	CreateRound(ctx context.Context, r *model.VoteRound, bookIDs []uuid.UUID) error
	GetOpenRound(ctx context.Context) (*model.VoteRound, error)
	GetRound(ctx context.Context, id uuid.UUID) (*model.VoteRound, error)
	ListRoundBooks(ctx context.Context, roundID uuid.UUID) ([]model.Book, error)

	CastVote(ctx context.Context, v *model.Vote) error
	GetVoteByMember(ctx context.Context, roundID, memberID uuid.UUID) (*model.Vote, error)
	CountVotes(ctx context.Context, roundID uuid.UUID) ([]model.VoteCount, error)

	CloseRound(ctx context.Context, id uuid.UUID, winner *uuid.UUID, closedAt time.Time) error
}

// ReviewStore — persistência de avaliações.
type ReviewStore interface {
	UpsertReview(ctx context.Context, r *model.Review) error
	GetReview(ctx context.Context, bookID, memberID uuid.UUID) (*model.Review, error)
	ListReviewsByBook(ctx context.Context, bookID uuid.UUID) ([]model.Review, error)
}

// MeetingStore — persistência de encontros, presenças e pauta.
type MeetingStore interface {
	CreateMeeting(ctx context.Context, m *model.Meeting) error
	GetMeeting(ctx context.Context, id uuid.UUID) (*model.Meeting, error)
	ListUpcoming(ctx context.Context, from time.Time) ([]model.Meeting, error)
	ListDueForReminder(ctx context.Context, now time.Time, horizon time.Duration) ([]model.Meeting, error)
	MarkReminded(ctx context.Context, id uuid.UUID, at time.Time) error

	SetAttendance(ctx context.Context, a *model.Attendance) error
	ListAttendances(ctx context.Context, meetingID uuid.UUID) ([]model.Attendance, error)
	ListAttendeesForReminder(ctx context.Context, meetingID uuid.UUID) ([]model.Member, error)

	AddAgendaItem(ctx context.Context, it *model.AgendaItem) error
	ListAgenda(ctx context.Context, meetingID uuid.UUID) ([]model.AgendaItem, error)
}
