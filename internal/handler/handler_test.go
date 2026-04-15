package handler

// Testes de integração básicos dos handlers de autenticação.
// A lógica de domínio já é coberta pelos testes unitários no pacote service.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/clube-do-livro/app/internal/middleware"
	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/service"
	"github.com/clube-do-livro/app/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// authStore satisfaz apenas a MemberStore — o suficiente para os testes
// de autenticação dos handlers.
type authStore struct {
	members  map[uuid.UUID]*model.Member
	byEmail  map[string]*model.Member
	sessions map[string]*model.Session
}

func newAuthStore() *authStore {
	return &authStore{
		members:  map[uuid.UUID]*model.Member{},
		byEmail:  map[string]*model.Member{},
		sessions: map[string]*model.Session{},
	}
}

func (s *authStore) CreateMember(_ context.Context, m *model.Member) error {
	s.members[m.ID] = m
	s.byEmail[strings.ToLower(m.Email)] = m
	return nil
}
func (s *authStore) GetMemberByEmail(_ context.Context, e string) (*model.Member, error) {
	m, ok := s.byEmail[strings.ToLower(e)]
	if !ok {
		return nil, store.ErrNotFound
	}
	return m, nil
}
func (s *authStore) GetMemberByID(_ context.Context, id uuid.UUID) (*model.Member, error) {
	m, ok := s.members[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return m, nil
}
func (s *authStore) ListMembers(context.Context) ([]model.Member, error) { return nil, nil }
func (s *authStore) CreateSession(_ context.Context, se *model.Session) error {
	s.sessions[se.Token] = se
	return nil
}
func (s *authStore) GetSessionByToken(_ context.Context, t string) (*model.Session, error) {
	se, ok := s.sessions[t]
	if !ok {
		return nil, store.ErrNotFound
	}
	return se, nil
}
func (s *authStore) DeleteSession(_ context.Context, t string) error { delete(s.sessions, t); return nil }
func (s *authStore) DeleteExpiredSessions(_ context.Context, now time.Time) error {
	for k, v := range s.sessions {
		if v.ExpiresAt.Before(now) {
			delete(s.sessions, k)
		}
	}
	return nil
}

func TestLoginLogoutFlow(t *testing.T) {
	ctx := context.Background()
	s := newAuthStore()

	// cria membro
	hash, _ := bcrypt.GenerateFromPassword([]byte("s3cret"), bcrypt.DefaultCost)
	m := &model.Member{
		ID: uuid.New(), Name: "Zé", Email: "ze@clube.local",
		PasswordHash: string(hash),
	}
	require.NoError(t, s.CreateMember(ctx, m))

	auth := service.NewAuth(s)
	h := &authHandler{d: Deps{Auth: auth}}

	// POST /login OK
	form := "email=ze@clube.local&password=s3cret"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.doLogin(rr, req)
	require.Equal(t, http.StatusFound, rr.Code)
	var token string
	for _, c := range rr.Result().Cookies() {
		if c.Name == middleware.CookieName {
			token = c.Value
		}
	}
	require.NotEmpty(t, token)

	// POST /login senha errada → 200 com flash
	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("email=ze@clube.local&password=errada"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	h.doLogin(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "Credenciais")

	// GET /logout → redirecionamento + cookie expirado
	req = httptest.NewRequest(http.MethodGet, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: token})
	rr = httptest.NewRecorder()
	h.doLogout(rr, req)
	require.Equal(t, http.StatusFound, rr.Code)
}

func TestRequireAuthRedirecionaSemCookie(t *testing.T) {
	s := newAuthStore()
	auth := service.NewAuth(s)

	mw := middleware.RequireAuth(auth)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusFound, rr.Code)
	require.Equal(t, "/login", rr.Header().Get("Location"))
}
