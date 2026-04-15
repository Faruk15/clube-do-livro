package store

import (
	"context"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
)

// --- MemberStore ---

func (p *PG) CreateMember(ctx context.Context, m *model.Member) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO members (id, name, email, password_hash, is_admin, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		m.ID, m.Name, m.Email, m.PasswordHash, m.IsAdmin, m.CreatedAt)
	return err
}

func (p *PG) GetMemberByEmail(ctx context.Context, email string) (*model.Member, error) {
	row := p.Pool.QueryRow(ctx,
		`SELECT id,name,email,password_hash,is_admin,created_at FROM members WHERE email=$1`, email)
	var m model.Member
	if err := row.Scan(&m.ID, &m.Name, &m.Email, &m.PasswordHash, &m.IsAdmin, &m.CreatedAt); err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (p *PG) GetMemberByID(ctx context.Context, id uuid.UUID) (*model.Member, error) {
	row := p.Pool.QueryRow(ctx,
		`SELECT id,name,email,password_hash,is_admin,created_at FROM members WHERE id=$1`, id)
	var m model.Member
	if err := row.Scan(&m.ID, &m.Name, &m.Email, &m.PasswordHash, &m.IsAdmin, &m.CreatedAt); err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (p *PG) ListMembers(ctx context.Context) ([]model.Member, error) {
	rows, err := p.Pool.Query(ctx,
		`SELECT id,name,email,password_hash,is_admin,created_at FROM members ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Member
	for rows.Next() {
		var m model.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.PasswordHash, &m.IsAdmin, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// --- SessionStore ---

func (p *PG) CreateSession(ctx context.Context, s *model.Session) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO sessions (id, member_id, token, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5)`,
		s.ID, s.MemberID, s.Token, s.ExpiresAt, s.CreatedAt)
	return err
}

func (p *PG) GetSessionByToken(ctx context.Context, token string) (*model.Session, error) {
	row := p.Pool.QueryRow(ctx,
		`SELECT id,member_id,token,expires_at,created_at FROM sessions WHERE token=$1`, token)
	var s model.Session
	if err := row.Scan(&s.ID, &s.MemberID, &s.Token, &s.ExpiresAt, &s.CreatedAt); err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (p *PG) DeleteSession(ctx context.Context, token string) error {
	_, err := p.Pool.Exec(ctx, `DELETE FROM sessions WHERE token=$1`, token)
	return err
}

func (p *PG) DeleteExpiredSessions(ctx context.Context, now time.Time) error {
	_, err := p.Pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < $1`, now)
	return err
}
