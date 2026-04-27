package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/store"
	"github.com/google/uuid"
)

var (
	ErrSemPermissao = errors.New("sem permissão para remover esta sugestão")
	ErrNaoEhSugestao = errors.New("apenas sugestões podem ser removidas")
)

// BookService — operações sobre o catálogo.
type BookService struct {
	Books store.BookStore
	Now   func() time.Time
}

func NewBook(b store.BookStore) *BookService {
	return &BookService{Books: b, Now: time.Now}
}

// Suggest cria uma sugestão de livro a partir dos dados já confirmados pelo membro.
func (s *BookService) Suggest(ctx context.Context, member *model.Member, b *model.Book) (*model.Book, error) {
	if b.Title == "" {
		return nil, errors.New("título obrigatório")
	}
	b.ID = uuid.New()
	b.Status = model.StatusSugerido
	b.SuggestedBy = &member.ID
	b.CreatedAt = s.Now()
	if err := s.Books.CreateBook(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarkLido encerra a leitura atual (status -> lido) e registra a data.
func (s *BookService) MarkLido(ctx context.Context, bookID uuid.UUID) error {
	t := s.Now()
	return s.Books.UpdateBookStatus(ctx, bookID, model.StatusLido, &t)
}

// AddTag normaliza e grava uma tag (evita duplicidade ignorando caixa).
func (s *BookService) AddTag(ctx context.Context, bookID uuid.UUID, tag string) error {
	tag = strings.ToLower(strings.TrimSpace(tag))
	if tag == "" {
		return errors.New("tag vazia")
	}
	return s.Books.AddTag(ctx, bookID, tag)
}

func (s *BookService) List(ctx context.Context, statusFilter string) ([]model.Book, error) {
	return s.Books.ListBooks(ctx, statusFilter)
}

func (s *BookService) ListFinished(ctx context.Context) ([]model.Book, error) {
	return s.Books.ListFinished(ctx)
}

func (s *BookService) Get(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	return s.Books.GetBook(ctx, id)
}

func (s *BookService) RemoveSuggestion(ctx context.Context, member *model.Member, bookID uuid.UUID) error {
	b, err := s.Books.GetBook(ctx, bookID)
	if err != nil {
		return err
	}
	if b.Status != model.StatusSugerido {
		return ErrNaoEhSugestao
	}
	if !member.IsAdmin && (b.SuggestedBy == nil || *b.SuggestedBy != member.ID) {
		return ErrSemPermissao
	}
	return s.Books.DeleteBook(ctx, bookID)
}
