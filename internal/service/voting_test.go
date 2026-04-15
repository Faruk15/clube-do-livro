package service

import (
	"context"
	"testing"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestVoting_AbrirExigeDoisLivros(t *testing.T) {
	v := NewVoting(newFakeVoting())
	_, err := v.OpenRound(context.Background(), []uuid.UUID{uuid.New()})
	require.Error(t, err)
}

func TestVoting_CastEConta(t *testing.T) {
	ctx := context.Background()
	fv := newFakeVoting()
	a, b := uuid.New(), uuid.New()
	fv.bookTitle[a] = "Alfa"
	fv.bookTitle[b] = "Beta"
	v := NewVoting(fv)

	r, err := v.OpenRound(ctx, []uuid.UUID{a, b})
	require.NoError(t, err)

	m1, m2, m3 := uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, v.Cast(ctx, r.ID, m1, a))
	require.NoError(t, v.Cast(ctx, r.ID, m2, a))
	require.NoError(t, v.Cast(ctx, r.ID, m3, b))

	// um membro troca o voto — não deve duplicar
	require.NoError(t, v.Cast(ctx, r.ID, m3, a))

	counts, err := v.Counts(ctx, r.ID)
	require.NoError(t, err)
	// Alfa: 3, Beta: 0
	require.Equal(t, 3, counts[0].Count)
	require.Equal(t, a, counts[0].BookID)
}

func TestVoting_LivroForaDaRodada(t *testing.T) {
	ctx := context.Background()
	fv := newFakeVoting()
	a, b := uuid.New(), uuid.New()
	fv.bookTitle[a] = "A"
	fv.bookTitle[b] = "B"
	v := NewVoting(fv)
	r, err := v.OpenRound(ctx, []uuid.UUID{a, b})
	require.NoError(t, err)

	err = v.Cast(ctx, r.ID, uuid.New(), uuid.New()) // livro inexistente
	require.ErrorIs(t, err, ErrLivroForaDaRodada)
}

func TestVoting_CloseElegeVencedor(t *testing.T) {
	ctx := context.Background()
	fv := newFakeVoting()
	a, b := uuid.New(), uuid.New()
	fv.bookTitle[a] = "Alfa"
	fv.bookTitle[b] = "Beta"
	v := NewVoting(fv)

	r, err := v.OpenRound(ctx, []uuid.UUID{a, b})
	require.NoError(t, err)
	require.NoError(t, v.Cast(ctx, r.ID, uuid.New(), b))
	require.NoError(t, v.Cast(ctx, r.ID, uuid.New(), b))
	require.NoError(t, v.Cast(ctx, r.ID, uuid.New(), a))

	counts, err := v.Close(ctx, r.ID)
	require.NoError(t, err)
	// Beta vence
	require.Equal(t, b, counts[0].BookID)

	got, _ := fv.GetRound(ctx, r.ID)
	require.Equal(t, model.RoundEncerrada, got.Status)
	require.NotNil(t, got.WinnerBookID)
	require.Equal(t, b, *got.WinnerBookID)
}

func TestVoting_EmpateVenceAlfabetico(t *testing.T) {
	ctx := context.Background()
	fv := newFakeVoting()
	a, b := uuid.New(), uuid.New()
	fv.bookTitle[a] = "Alfa"
	fv.bookTitle[b] = "Beta"
	v := NewVoting(fv)
	r, err := v.OpenRound(ctx, []uuid.UUID{a, b})
	require.NoError(t, err)
	require.NoError(t, v.Cast(ctx, r.ID, uuid.New(), a))
	require.NoError(t, v.Cast(ctx, r.ID, uuid.New(), b))

	_, err = v.Close(ctx, r.ID)
	require.NoError(t, err)

	got, _ := fv.GetRound(ctx, r.ID)
	require.NotNil(t, got.WinnerBookID)
	require.Equal(t, a, *got.WinnerBookID)
}

func TestVoting_CloseSemVotos(t *testing.T) {
	ctx := context.Background()
	fv := newFakeVoting()
	a, b := uuid.New(), uuid.New()
	fv.bookTitle[a] = "A"
	fv.bookTitle[b] = "B"
	v := NewVoting(fv)
	r, _ := v.OpenRound(ctx, []uuid.UUID{a, b})
	_, err := v.Close(ctx, r.ID)
	require.NoError(t, err)
	got, _ := fv.GetRound(ctx, r.ID)
	require.Nil(t, got.WinnerBookID)
}

func TestVoting_CastEmRodadaEncerrada(t *testing.T) {
	ctx := context.Background()
	fv := newFakeVoting()
	a, b := uuid.New(), uuid.New()
	fv.bookTitle[a] = "A"
	fv.bookTitle[b] = "B"
	v := NewVoting(fv)
	r, _ := v.OpenRound(ctx, []uuid.UUID{a, b})
	_, _ = v.Close(ctx, r.ID)

	err := v.Cast(ctx, r.ID, uuid.New(), a)
	require.ErrorIs(t, err, ErrRodadaEncerrada)
}
