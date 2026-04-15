package store

import (
	"context"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
)

func (p *PG) CreateBook(ctx context.Context, b *model.Book) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	if b.Status == "" {
		b.Status = model.StatusSugerido
	}
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO books (id,title,author,cover_url,synopsis,publisher,year,pages,status,suggested_by,finished_at,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		b.ID, b.Title, b.Author, b.CoverURL, b.Synopsis, b.Publisher, b.Year, b.Pages,
		b.Status, b.SuggestedBy, b.FinishedAt, b.CreatedAt)
	return err
}

func (p *PG) GetBook(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	row := p.Pool.QueryRow(ctx, `
		SELECT id,title,author,cover_url,synopsis,publisher,year,pages,status,suggested_by,finished_at,created_at
		FROM books WHERE id=$1`, id)
	b, err := scanBook(row)
	if err != nil {
		return nil, err
	}
	tags, err := p.ListTags(ctx, id)
	if err != nil {
		return nil, err
	}
	b.Tags = tags
	return b, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBook(r rowScanner) (*model.Book, error) {
	var b model.Book
	err := r.Scan(&b.ID, &b.Title, &b.Author, &b.CoverURL, &b.Synopsis, &b.Publisher,
		&b.Year, &b.Pages, &b.Status, &b.SuggestedBy, &b.FinishedAt, &b.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (p *PG) UpdateBookStatus(ctx context.Context, id uuid.UUID, status string, finishedAt *time.Time) error {
	_, err := p.Pool.Exec(ctx,
		`UPDATE books SET status=$2, finished_at=$3 WHERE id=$1`, id, status, finishedAt)
	return err
}

func (p *PG) ListBooks(ctx context.Context, statusFilter string) ([]model.Book, error) {
	query := `SELECT id,title,author,cover_url,synopsis,publisher,year,pages,status,suggested_by,finished_at,created_at
	          FROM books`
	args := []any{}
	if statusFilter != "" {
		query += ` WHERE status=$1`
		args = append(args, statusFilter)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := p.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Book
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *b)
	}
	return out, rows.Err()
}

func (p *PG) ListFinished(ctx context.Context) ([]model.Book, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT id,title,author,cover_url,synopsis,publisher,year,pages,status,suggested_by,finished_at,created_at
		FROM books WHERE status='lido' ORDER BY finished_at DESC NULLS LAST`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Book
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *b)
	}
	return out, rows.Err()
}

func (p *PG) AddTag(ctx context.Context, bookID uuid.UUID, tag string) error {
	_, err := p.Pool.Exec(ctx,
		`INSERT INTO book_tags(book_id, tag) VALUES($1,$2) ON CONFLICT DO NOTHING`, bookID, tag)
	return err
}

func (p *PG) RemoveTag(ctx context.Context, bookID uuid.UUID, tag string) error {
	_, err := p.Pool.Exec(ctx, `DELETE FROM book_tags WHERE book_id=$1 AND tag=$2`, bookID, tag)
	return err
}

func (p *PG) ListTags(ctx context.Context, bookID uuid.UUID) ([]string, error) {
	rows, err := p.Pool.Query(ctx, `SELECT tag FROM book_tags WHERE book_id=$1 ORDER BY tag`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
