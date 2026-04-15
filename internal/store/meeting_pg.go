package store

import (
	"context"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
)

func (p *PG) CreateMeeting(ctx context.Context, m *model.Meeting) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO meetings(id,title,datetime,location,book_id,reminded_at,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7)`,
		m.ID, m.Title, m.Datetime, m.Location, m.BookID, m.RemindedAt, m.CreatedAt)
	return err
}

func (p *PG) GetMeeting(ctx context.Context, id uuid.UUID) (*model.Meeting, error) {
	row := p.Pool.QueryRow(ctx,
		`SELECT id,title,datetime,location,book_id,reminded_at,created_at FROM meetings WHERE id=$1`, id)
	var m model.Meeting
	if err := row.Scan(&m.ID, &m.Title, &m.Datetime, &m.Location, &m.BookID, &m.RemindedAt, &m.CreatedAt); err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (p *PG) ListUpcoming(ctx context.Context, from time.Time) ([]model.Meeting, error) {
	rows, err := p.Pool.Query(ctx,
		`SELECT id,title,datetime,location,book_id,reminded_at,created_at
		 FROM meetings WHERE datetime >= $1 ORDER BY datetime ASC`, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Meeting
	for rows.Next() {
		var m model.Meeting
		if err := rows.Scan(&m.ID, &m.Title, &m.Datetime, &m.Location, &m.BookID, &m.RemindedAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (p *PG) ListDueForReminder(ctx context.Context, now time.Time, horizon time.Duration) ([]model.Meeting, error) {
	rows, err := p.Pool.Query(ctx,
		`SELECT id,title,datetime,location,book_id,reminded_at,created_at
		 FROM meetings
		 WHERE reminded_at IS NULL
		   AND datetime BETWEEN $1 AND $2`,
		now, now.Add(horizon))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Meeting
	for rows.Next() {
		var m model.Meeting
		if err := rows.Scan(&m.ID, &m.Title, &m.Datetime, &m.Location, &m.BookID, &m.RemindedAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (p *PG) MarkReminded(ctx context.Context, id uuid.UUID, at time.Time) error {
	_, err := p.Pool.Exec(ctx, `UPDATE meetings SET reminded_at=$2 WHERE id=$1`, id, at)
	return err
}

func (p *PG) SetAttendance(ctx context.Context, a *model.Attendance) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO attendances(id,meeting_id,member_id,status)
		VALUES($1,$2,$3,$4)
		ON CONFLICT (meeting_id, member_id) DO UPDATE SET status=EXCLUDED.status`,
		a.ID, a.MeetingID, a.MemberID, a.Status)
	return err
}

func (p *PG) ListAttendances(ctx context.Context, meetingID uuid.UUID) ([]model.Attendance, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT a.id,a.meeting_id,a.member_id,a.status,m.name
		FROM attendances a JOIN members m ON m.id = a.member_id
		WHERE a.meeting_id=$1 ORDER BY m.name`, meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Attendance
	for rows.Next() {
		var a model.Attendance
		if err := rows.Scan(&a.ID, &a.MeetingID, &a.MemberID, &a.Status, &a.Name); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (p *PG) ListAttendeesForReminder(ctx context.Context, meetingID uuid.UUID) ([]model.Member, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT m.id,m.name,m.email,m.password_hash,m.is_admin,m.created_at
		FROM members m
		JOIN attendances a ON a.member_id = m.id
		WHERE a.meeting_id=$1 AND a.status IN ('confirmado','talvez')`, meetingID)
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

func (p *PG) AddAgendaItem(ctx context.Context, it *model.AgendaItem) error {
	if it.ID == uuid.Nil {
		it.ID = uuid.New()
	}
	if it.CreatedAt.IsZero() {
		it.CreatedAt = time.Now()
	}
	_, err := p.Pool.Exec(ctx,
		`INSERT INTO agenda_items(id,meeting_id,member_id,content,created_at) VALUES($1,$2,$3,$4,$5)`,
		it.ID, it.MeetingID, it.MemberID, it.Content, it.CreatedAt)
	return err
}

func (p *PG) ListAgenda(ctx context.Context, meetingID uuid.UUID) ([]model.AgendaItem, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT a.id,a.meeting_id,a.member_id,a.content,a.created_at,m.name
		FROM agenda_items a JOIN members m ON m.id = a.member_id
		WHERE a.meeting_id=$1 ORDER BY a.created_at ASC`, meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AgendaItem
	for rows.Next() {
		var it model.AgendaItem
		if err := rows.Scan(&it.ID, &it.MeetingID, &it.MemberID, &it.Content, &it.CreatedAt, &it.Author); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
