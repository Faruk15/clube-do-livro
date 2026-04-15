package store

import (
	"context"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
)

func (p *PG) CreateRound(ctx context.Context, r *model.VoteRound, bookIDs []uuid.UUID) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.OpenedAt.IsZero() {
		r.OpenedAt = time.Now()
	}
	if r.Status == "" {
		r.Status = model.RoundAberta
	}

	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO vote_rounds(id,status,opened_at) VALUES($1,$2,$3)`,
		r.ID, r.Status, r.OpenedAt)
	if err != nil {
		return err
	}
	for _, bID := range bookIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO vote_round_books(round_id, book_id) VALUES($1,$2)`, r.ID, bID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE books SET status=$2 WHERE id=$1`, bID, model.StatusEmVotacao); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (p *PG) GetOpenRound(ctx context.Context) (*model.VoteRound, error) {
	row := p.Pool.QueryRow(ctx,
		`SELECT id,status,opened_at,closed_at,winner_book_id FROM vote_rounds
		 WHERE status='aberta' ORDER BY opened_at DESC LIMIT 1`)
	var r model.VoteRound
	if err := row.Scan(&r.ID, &r.Status, &r.OpenedAt, &r.ClosedAt, &r.WinnerBookID); err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (p *PG) GetRound(ctx context.Context, id uuid.UUID) (*model.VoteRound, error) {
	row := p.Pool.QueryRow(ctx,
		`SELECT id,status,opened_at,closed_at,winner_book_id FROM vote_rounds WHERE id=$1`, id)
	var r model.VoteRound
	if err := row.Scan(&r.ID, &r.Status, &r.OpenedAt, &r.ClosedAt, &r.WinnerBookID); err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (p *PG) ListRoundBooks(ctx context.Context, roundID uuid.UUID) ([]model.Book, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT b.id,b.title,b.author,b.cover_url,b.synopsis,b.publisher,b.year,b.pages,
		       b.status,b.suggested_by,b.finished_at,b.created_at
		FROM books b
		JOIN vote_round_books vrb ON vrb.book_id = b.id
		WHERE vrb.round_id = $1
		ORDER BY b.title`, roundID)
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

func (p *PG) CastVote(ctx context.Context, v *model.Vote) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}
	// ON CONFLICT permite ao membro trocar seu voto enquanto a rodada está aberta.
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO votes(id,round_id,member_id,book_id,created_at)
		VALUES($1,$2,$3,$4,$5)
		ON CONFLICT (round_id, member_id) DO UPDATE
		SET book_id = EXCLUDED.book_id, created_at = EXCLUDED.created_at`,
		v.ID, v.RoundID, v.MemberID, v.BookID, v.CreatedAt)
	return err
}

func (p *PG) GetVoteByMember(ctx context.Context, roundID, memberID uuid.UUID) (*model.Vote, error) {
	row := p.Pool.QueryRow(ctx,
		`SELECT id,round_id,member_id,book_id,created_at FROM votes
		 WHERE round_id=$1 AND member_id=$2`, roundID, memberID)
	var v model.Vote
	if err := row.Scan(&v.ID, &v.RoundID, &v.MemberID, &v.BookID, &v.CreatedAt); err != nil {
		if isNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (p *PG) CountVotes(ctx context.Context, roundID uuid.UUID) ([]model.VoteCount, error) {
	rows, err := p.Pool.Query(ctx, `
		SELECT b.id, b.title, COUNT(v.id) AS total
		FROM vote_round_books vrb
		JOIN books b ON b.id = vrb.book_id
		LEFT JOIN votes v ON v.book_id = b.id AND v.round_id = vrb.round_id
		WHERE vrb.round_id = $1
		GROUP BY b.id, b.title
		ORDER BY total DESC, b.title`, roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.VoteCount
	for rows.Next() {
		var c model.VoteCount
		if err := rows.Scan(&c.BookID, &c.Title, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (p *PG) CloseRound(ctx context.Context, id uuid.UUID, winner *uuid.UUID, closedAt time.Time) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE vote_rounds SET status='encerrada', closed_at=$2, winner_book_id=$3 WHERE id=$1`,
		id, closedAt, winner)
	if err != nil {
		return err
	}

	// Livros em_votacao voltam para sugerido, exceto o vencedor que vira lendo_agora.
	_, err = tx.Exec(ctx, `
		UPDATE books SET status='sugerido'
		WHERE id IN (SELECT book_id FROM vote_round_books WHERE round_id=$1)
		  AND status='em_votacao'`, id)
	if err != nil {
		return err
	}
	if winner != nil {
		_, err = tx.Exec(ctx, `UPDATE books SET status='lendo_agora' WHERE id=$1`, *winner)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
