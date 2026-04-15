// Package service concentra a lógica de negócio do clube.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/store"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Erros de domínio (o handler mapeia em HTTP).
var (
	ErrCredenciaisInvalidas = errors.New("credenciais inválidas")
	ErrEmailJaExiste        = errors.New("e-mail já cadastrado")
	ErrNaoAutorizado        = errors.New("não autorizado")
)

// SessionDuration: quanto tempo uma sessão dura antes de expirar.
const SessionDuration = 30 * 24 * time.Hour

// AuthService — cuida de login, logout e criação de membros.
type AuthService struct {
	Members store.MemberStore
	Now     func() time.Time
}

func NewAuth(m store.MemberStore) *AuthService {
	return &AuthService{Members: m, Now: time.Now}
}

// CreateMember cria um novo membro. Retorna ErrEmailJaExiste se já houver.
func (a *AuthService) CreateMember(ctx context.Context, name, email, password string, isAdmin bool) (*model.Member, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" || password == "" {
		return nil, errors.New("nome, e-mail e senha são obrigatórios")
	}
	if _, err := a.Members.GetMemberByEmail(ctx, email); err == nil {
		return nil, ErrEmailJaExiste
	} else if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	m := &model.Member{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		IsAdmin:      isAdmin,
		CreatedAt:    a.Now(),
	}
	if err := a.Members.CreateMember(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// Login valida e-mail+senha, cria uma sessão e devolve o token opaco a ser
// gravado no cookie. Ignora caixa e espaços no e-mail.
func (a *AuthService) Login(ctx context.Context, email, password string) (string, *model.Member, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	m, err := a.Members.GetMemberByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return "", nil, ErrCredenciaisInvalidas
		}
		return "", nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(m.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrCredenciaisInvalidas
	}
	token, err := randomToken()
	if err != nil {
		return "", nil, err
	}
	sess := &model.Session{
		ID:        uuid.New(),
		MemberID:  m.ID,
		Token:     token,
		ExpiresAt: a.Now().Add(SessionDuration),
		CreatedAt: a.Now(),
	}
	if err := a.Members.CreateSession(ctx, sess); err != nil {
		return "", nil, err
	}
	return token, m, nil
}

// Logout apaga a sessão vinculada ao token. Idempotente.
func (a *AuthService) Logout(ctx context.Context, token string) error {
	return a.Members.DeleteSession(ctx, token)
}

// MemberFromToken resolve o token de cookie em um membro autenticado.
// Retorna ErrNaoAutorizado se a sessão não existir ou estiver expirada.
func (a *AuthService) MemberFromToken(ctx context.Context, token string) (*model.Member, error) {
	if token == "" {
		return nil, ErrNaoAutorizado
	}
	s, err := a.Members.GetSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrNaoAutorizado
		}
		return nil, err
	}
	if s.ExpiresAt.Before(a.Now()) {
		_ = a.Members.DeleteSession(ctx, token)
		return nil, ErrNaoAutorizado
	}
	return a.Members.GetMemberByID(ctx, s.MemberID)
}

// randomToken gera um token opaco hex (32 bytes => 64 chars).
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
