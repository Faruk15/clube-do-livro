package store

import (
	"context"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
)

func (p *PG) UpsertReview(ctx context.Context, r *model.Review) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	now := time.Now()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	r.UpdatedAt = now

	_, err := p.Pool.Exec(ctx, `
		INSERT INTO reviews(id,book_id,member_id,nota_geral,nota_escrita,nota_enredo,nota_expectativa,
		                    review_text,has_spoiler,citacao,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (book_id, member_id) DO UPDATE SET
		    nota_geral=EXCLUDED.nota_geral,
		    nota_escrita=EXCLUDED.nota_escrita,
		    nota_enredo=EXCLUDED.nota_enredo,
		    nota_expectativa=EXCLUDED.nota_expectativa,
		    review_text=EXCLUDED.review_text,
		    has_spoiler=EXCLUDED.has_spoiler,
		    citacao=EXCLUDED.citacao,
		    updated_at=EXCLUDED.updated_at`,
		r.ID, r.BookID, r.MemberID, r.NotaGeral, r.NotaEscrita, r.NotaEnredo, r.NotaExpectativa,
		r.ReviewText, r.HasSpoiler, r.Citacao, r.CreatedAt, r.UpdatedAt)
	return err
}

func (p *PG) GetReview(ctx context.Context, bookID, memberID uuid.UUID) (*model.Review, error) {
	row := p.Pool.QueryRow(ctx, `
		SELECT id,book_id,member_id,nota_geral,nota_escrita,nota_enredo,nota_expectativa,
		       review_text,has_spoiler,citacao,created_at,updated_at
		FROM reviews WHERE book_id=$1 AND member_id=$2`, bookID, memberID)
	var r model.Review
	err := row.Scan(&r.ID, &r.BookID, &r.MemberID, &r.NotaGeral, &r.NotaEscrita, &r.NotaEnredo,
		&r.NotaExpectativa, &r.ReviewText, &r.HasSpoiler, &r.Citacao, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (p *PG) ListReviewsByBook(ctx context.Context, bookID uuid.UUID) ([]model.Review, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT id,book_id,member_id,nota_geral,nota_escrita,nota_enredo,nota_expectativa,
		       review_text,has_spoiler,citacao,created_at,updated_at
		FROM reviews WHERE book_id=$1 ORDER BY created_at`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Review
	for rows.Next() {
		var r model.Review
		if err := rows.Scan(&r.ID, &r.BookID, &r.MemberID, &r.NotaGeral, &r.NotaEscrita, &r.NotaEnredo,
			&r.NotaExpectativa, &r.ReviewText, &r.HasSpoiler, &r.Citacao, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
