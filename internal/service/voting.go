package service

import (
	"context"
	"errors"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/store"
	"github.com/google/uuid"
)

var (
	ErrRodadaInexistente = errors.New("nenhuma rodada aberta")
	ErrRodadaEncerrada   = errors.New("rodada encerrada")
	ErrLivroForaDaRodada = errors.New("livro não está nesta rodada")
)

// VotingService — abertura, voto e encerramento de rodadas.
type VotingService struct {
	Rounds store.VotingStore
	Now    func() time.Time
}

func NewVoting(r store.VotingStore) *VotingService {
	return &VotingService{Rounds: r, Now: time.Now}
}

// OpenRound cria uma rodada com os livros informados (que devem estar sugeridos).
// O store marca automaticamente os livros como em_votacao.
func (v *VotingService) OpenRound(ctx context.Context, bookIDs []uuid.UUID) (*model.VoteRound, error) {
	if len(bookIDs) < 2 {
		return nil, errors.New("ao menos 2 livros são necessários para abrir votação")
	}
	r := &model.VoteRound{
		ID:       uuid.New(),
		Status:   model.RoundAberta,
		OpenedAt: v.Now(),
	}
	if err := v.Rounds.CreateRound(ctx, r, bookIDs); err != nil {
		return nil, err
	}
	return r, nil
}

// Cast registra (ou troca) o voto do membro. Impede voto após encerramento e
// força que o livro seja um dos candidatos da rodada.
func (v *VotingService) Cast(ctx context.Context, roundID, memberID, bookID uuid.UUID) error {
	r, err := v.Rounds.GetRound(ctx, roundID)
	if err != nil {
		return err
	}
	if r.Status != model.RoundAberta {
		return ErrRodadaEncerrada
	}
	books, err := v.Rounds.ListRoundBooks(ctx, roundID)
	if err != nil {
		return err
	}
	if !containsBook(books, bookID) {
		return ErrLivroForaDaRodada
	}
	return v.Rounds.CastVote(ctx, &model.Vote{
		RoundID:  roundID,
		MemberID: memberID,
		BookID:   bookID,
	})
}

// Close encerra a rodada, apura o vencedor e ajusta o status dos livros.
// Empate: o primeiro livro em ordem alfabética vence (comportamento determinístico).
func (v *VotingService) Close(ctx context.Context, roundID uuid.UUID) ([]model.VoteCount, error) {
	r, err := v.Rounds.GetRound(ctx, roundID)
	if err != nil {
		return nil, err
	}
	if r.Status != model.RoundAberta {
		return nil, ErrRodadaEncerrada
	}
	counts, err := v.Rounds.CountVotes(ctx, roundID)
	if err != nil {
		return nil, err
	}
	winner := pickWinner(counts)
	if err := v.Rounds.CloseRound(ctx, roundID, winner, v.Now()); err != nil {
		return nil, err
	}
	return counts, nil
}

// CurrentRound devolve a rodada aberta ou nil.
func (v *VotingService) CurrentRound(ctx context.Context) (*model.VoteRound, []model.Book, error) {
	r, err := v.Rounds.GetOpenRound(ctx)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	books, err := v.Rounds.ListRoundBooks(ctx, r.ID)
	if err != nil {
		return nil, nil, err
	}
	return r, books, nil
}

// MyVote retorna o voto atual do membro na rodada (ou nil).
func (v *VotingService) MyVote(ctx context.Context, roundID, memberID uuid.UUID) (*model.Vote, error) {
	vt, err := v.Rounds.GetVoteByMember(ctx, roundID, memberID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	return vt, err
}

// Counts retorna a apuração de uma rodada (usado em listagens admin).
func (v *VotingService) Counts(ctx context.Context, roundID uuid.UUID) ([]model.VoteCount, error) {
	return v.Rounds.CountVotes(ctx, roundID)
}

func containsBook(books []model.Book, id uuid.UUID) bool {
	for _, b := range books {
		if b.ID == id {
			return true
		}
	}
	return false
}

// pickWinner: livro com mais votos; em caso de empate, o primeiro alfabético.
// O store já devolve ordenado por total DESC, título ASC — então basta pegar o topo
// se houver algum voto.
func pickWinner(counts []model.VoteCount) *uuid.UUID {
	if len(counts) == 0 {
		return nil
	}
	top := counts[0]
	if top.Count == 0 {
		return nil
	}
	id := top.BookID
	return &id
}
