package service

import (
	"context"
	"errors"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/store"
	"github.com/google/uuid"
)

// ReviewService — avalia apenas livros com status "lido". Todos os
// campos numéricos são opcionais e nil significa "não respondeu".
type ReviewService struct {
	Reviews store.ReviewStore
	Books   store.BookStore
}

func NewReview(r store.ReviewStore, b store.BookStore) *ReviewService {
	return &ReviewService{Reviews: r, Books: b}
}

// ErrLivroNaoLido — só livros lidos podem ser avaliados.
var ErrLivroNaoLido = errors.New("livro ainda não foi marcado como lido")

// Upsert cria ou atualiza a avaliação do membro. Valida o range das notas (1..5) se presentes.
func (s *ReviewService) Upsert(ctx context.Context, memberID uuid.UUID, r *model.Review) error {
	book, err := s.Books.GetBook(ctx, r.BookID)
	if err != nil {
		return err
	}
	if book.Status != model.StatusLido {
		return ErrLivroNaoLido
	}
	if err := validNota(r.NotaGeral); err != nil {
		return err
	}
	if err := validNota(r.NotaEscrita); err != nil {
		return err
	}
	if err := validNota(r.NotaEnredo); err != nil {
		return err
	}
	if err := validNota(r.NotaExpectativa); err != nil {
		return err
	}
	r.MemberID = memberID
	return s.Reviews.UpsertReview(ctx, r)
}

func validNota(n *int) error {
	if n == nil {
		return nil
	}
	if *n < 1 || *n > 5 {
		return errors.New("nota fora do intervalo 1..5")
	}
	return nil
}

// Stats calcula a média apenas sobre os campos efetivamente preenchidos por cada
// avaliador. Retorna zero em campos sem nenhuma resposta.
func (s *ReviewService) Stats(ctx context.Context, bookID uuid.UUID) (model.ReviewStats, []model.Review, error) {
	revs, err := s.Reviews.ListReviewsByBook(ctx, bookID)
	if err != nil {
		return model.ReviewStats{}, nil, err
	}
	st := model.ReviewStats{BookID: bookID, TotalReviews: len(revs)}
	var sumG, sumE, sumEn, sumX int
	for _, r := range revs {
		if r.NotaGeral != nil {
			sumG += *r.NotaGeral
			st.CountGeral++
		}
		if r.NotaEscrita != nil {
			sumE += *r.NotaEscrita
			st.CountEscrita++
		}
		if r.NotaEnredo != nil {
			sumEn += *r.NotaEnredo
			st.CountEnredo++
		}
		if r.NotaExpectativa != nil {
			sumX += *r.NotaExpectativa
			st.CountExpectativa++
		}
	}
	if st.CountGeral > 0 {
		st.AvgGeral = float64(sumG) / float64(st.CountGeral)
	}
	if st.CountEscrita > 0 {
		st.AvgEscrita = float64(sumE) / float64(st.CountEscrita)
	}
	if st.CountEnredo > 0 {
		st.AvgEnredo = float64(sumEn) / float64(st.CountEnredo)
	}
	if st.CountExpectativa > 0 {
		st.AvgExpectativa = float64(sumX) / float64(st.CountExpectativa)
	}
	return st, revs, nil
}

// Mine devolve a avaliação do membro, ou nil se ainda não fez.
func (s *ReviewService) Mine(ctx context.Context, bookID, memberID uuid.UUID) (*model.Review, error) {
	r, err := s.Reviews.GetReview(ctx, bookID, memberID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	return r, err
}
