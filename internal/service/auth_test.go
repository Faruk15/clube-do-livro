package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuth_CreateAndLogin(t *testing.T) {
	ctx := context.Background()
	a := NewAuth(newFakeMembers())

	m, err := a.CreateMember(ctx, "Ana", "ana@clube.local", "senha123", false)
	require.NoError(t, err)
	require.Equal(t, "ana@clube.local", m.Email)

	token, got, err := a.Login(ctx, "ANA@clube.local", "senha123")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.Equal(t, m.ID, got.ID)

	// token resolve de volta para o membro
	me, err := a.MemberFromToken(ctx, token)
	require.NoError(t, err)
	require.Equal(t, m.ID, me.ID)
}

func TestAuth_LoginSenhaErrada(t *testing.T) {
	ctx := context.Background()
	a := NewAuth(newFakeMembers())
	_, err := a.CreateMember(ctx, "Ana", "ana@clube.local", "senha123", false)
	require.NoError(t, err)

	_, _, err = a.Login(ctx, "ana@clube.local", "errada")
	require.ErrorIs(t, err, ErrCredenciaisInvalidas)
}

func TestAuth_CreateEmailDuplicado(t *testing.T) {
	ctx := context.Background()
	a := NewAuth(newFakeMembers())
	_, err := a.CreateMember(ctx, "Ana", "ana@clube.local", "senha123", false)
	require.NoError(t, err)
	_, err = a.CreateMember(ctx, "Outra Ana", "ana@clube.local", "s2", false)
	require.ErrorIs(t, err, ErrEmailJaExiste)
}

func TestAuth_Logout(t *testing.T) {
	ctx := context.Background()
	a := NewAuth(newFakeMembers())
	_, err := a.CreateMember(ctx, "Ana", "ana@clube.local", "senha123", false)
	require.NoError(t, err)
	tok, _, err := a.Login(ctx, "ana@clube.local", "senha123")
	require.NoError(t, err)

	require.NoError(t, a.Logout(ctx, tok))
	_, err = a.MemberFromToken(ctx, tok)
	require.ErrorIs(t, err, ErrNaoAutorizado)
}

func TestAuth_SessaoExpirada(t *testing.T) {
	ctx := context.Background()
	fm := newFakeMembers()
	a := NewAuth(fm)
	a.Now = func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }
	_, err := a.CreateMember(ctx, "Ana", "ana@clube.local", "senha123", false)
	require.NoError(t, err)

	tok, _, err := a.Login(ctx, "ana@clube.local", "senha123")
	require.NoError(t, err)

	// avança o relógio além da expiração
	a.Now = func() time.Time { return time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC) }
	_, err = a.MemberFromToken(ctx, tok)
	require.ErrorIs(t, err, ErrNaoAutorizado)
}
