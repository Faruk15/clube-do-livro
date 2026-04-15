package service

import (
	"context"
	"testing"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func intPtr(i int) *int { return &i }

func seedLido(t *testing.T, fb *fakeBooks) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, fb.CreateBook(context.Background(), &model.Book{ID: id, Title: "L", Status: model.StatusLido}))
	return id
}

func TestReview_UpsertExigeLido(t *testing.T) {
	ctx := context.Background()
	fr := newFakeReviews()
	fb := newFakeBooks()
	// livro não está "lido"
	id := uuid.New()
	require.NoError(t, fb.CreateBook(ctx, &model.Book{ID: id, Title: "X", Status: model.StatusSugerido}))

	s := NewReview(fr, fb)
	err := s.Upsert(ctx, uuid.New(), &model.Review{BookID: id, NotaGeral: intPtr(5)})
	require.ErrorIs(t, err, ErrLivroNaoLido)
}

func TestReview_UpsertValidaRange(t *testing.T) {
	ctx := context.Background()
	fr := newFakeReviews()
	fb := newFakeBooks()
	id := seedLido(t, fb)
	s := NewReview(fr, fb)
	err := s.Upsert(ctx, uuid.New(), &model.Review{BookID: id, NotaGeral: intPtr(9)})
	require.Error(t, err)
}

func TestReview_StatsMediaIgnoraNulos(t *testing.T) {
	ctx := context.Background()
	fr := newFakeReviews()
	fb := newFakeBooks()
	id := seedLido(t, fb)
	s := NewReview(fr, fb)

	// 3 avaliações, cobrindo parcialmente os campos
	require.NoError(t, s.Upsert(ctx, uuid.New(), &model.Review{BookID: id, NotaGeral: intPtr(5), NotaEscrita: intPtr(4)}))
	require.NoError(t, s.Upsert(ctx, uuid.New(), &model.Review{BookID: id, NotaGeral: intPtr(3)}))
	require.NoError(t, s.Upsert(ctx, uuid.New(), &model.Review{BookID: id, NotaEnredo: intPtr(5), HasSpoiler: true, ReviewText: "spoiler!"}))

	stats, revs, err := s.Stats(ctx, id)
	require.NoError(t, err)
	require.Len(t, revs, 3)
	require.Equal(t, 2, stats.CountGeral)
	require.InDelta(t, 4.0, stats.AvgGeral, 0.001)
	require.Equal(t, 1, stats.CountEscrita)
	require.InDelta(t, 4.0, stats.AvgEscrita, 0.001)
	require.Equal(t, 1, stats.CountEnredo)
	require.InDelta(t, 5.0, stats.AvgEnredo, 0.001)
	require.Equal(t, 0, stats.CountExpectativa)
	require.Equal(t, 0.0, stats.AvgExpectativa)
}

func TestReview_UpsertAtualiza(t *testing.T) {
	ctx := context.Background()
	fr := newFakeReviews()
	fb := newFakeBooks()
	id := seedLido(t, fb)
	s := NewReview(fr, fb)
	member := uuid.New()

	require.NoError(t, s.Upsert(ctx, member, &model.Review{BookID: id, NotaGeral: intPtr(3)}))
	require.NoError(t, s.Upsert(ctx, member, &model.Review{BookID: id, NotaGeral: intPtr(5)}))

	r, err := s.Mine(ctx, id, member)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, 5, *r.NotaGeral)

	_, revs, _ := s.Stats(ctx, id)
	require.Len(t, revs, 1) // não duplicou
}
