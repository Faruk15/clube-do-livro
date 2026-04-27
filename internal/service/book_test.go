package service

import (
	"context"
	"testing"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/stretchr/testify/require"
)

func TestBook_Suggest(t *testing.T) {
	ctx := context.Background()
	fb := newFakeBooks()
	s := NewBook(fb)

	member := &model.Member{ID: mustUUID()}
	b, err := s.Suggest(ctx, member, &model.Book{Title: "Novo"})
	require.NoError(t, err)
	require.Equal(t, model.StatusSugerido, b.Status)
	require.NotNil(t, b.SuggestedBy)
	require.Equal(t, member.ID, *b.SuggestedBy)
}

func TestBook_SuggestExigeTitulo(t *testing.T) {
	fb := newFakeBooks()
	s := NewBook(fb)
	_, err := s.Suggest(context.Background(), &model.Member{ID: mustUUID()}, &model.Book{})
	require.Error(t, err)
}

func TestBook_RemoveSuggestion(t *testing.T) {
	ctx := context.Background()
	fb := newFakeBooks()
	s := NewBook(fb)

	owner := &model.Member{ID: mustUUID()}
	other := &model.Member{ID: mustUUID()}
	admin := &model.Member{ID: mustUUID(), IsAdmin: true}

	b, _ := s.Suggest(ctx, owner, &model.Book{Title: "Remover"})

	// outro membro não pode remover
	require.ErrorIs(t, s.RemoveSuggestion(ctx, other, b.ID), ErrSemPermissao)

	// admin pode remover qualquer sugestão
	b2, _ := s.Suggest(ctx, owner, &model.Book{Title: "Admin remove"})
	require.NoError(t, s.RemoveSuggestion(ctx, admin, b2.ID))
	_, err := fb.GetBook(ctx, b2.ID)
	require.Error(t, err)

	// dono pode remover a própria sugestão
	require.NoError(t, s.RemoveSuggestion(ctx, owner, b.ID))
	_, err = fb.GetBook(ctx, b.ID)
	require.Error(t, err)
}

func TestBook_AddTagNormaliza(t *testing.T) {
	ctx := context.Background()
	fb := newFakeBooks()
	s := NewBook(fb)
	b, _ := s.Suggest(ctx, &model.Member{ID: mustUUID()}, &model.Book{Title: "x"})
	require.NoError(t, s.AddTag(ctx, b.ID, "  Ficção  "))
	tags, _ := fb.ListTags(ctx, b.ID)
	require.Contains(t, tags, "ficção")
}
