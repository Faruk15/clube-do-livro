package service

// Fakes in-memory para os testes do service. Implementam as interfaces
// do pacote store/ de forma bem simples, apenas o suficiente para os
// cenários cobertos pelos testes.

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/store"
	"github.com/google/uuid"
)

// --- MemberStore fake ---

type fakeMembers struct {
	mu       sync.Mutex
	members  map[uuid.UUID]*model.Member
	byEmail  map[string]*model.Member
	sessions map[string]*model.Session
}

func newFakeMembers() *fakeMembers {
	return &fakeMembers{
		members:  map[uuid.UUID]*model.Member{},
		byEmail:  map[string]*model.Member{},
		sessions: map[string]*model.Session{},
	}
}

func (f *fakeMembers) CreateMember(_ context.Context, m *model.Member) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.members[m.ID] = m
	f.byEmail[strings.ToLower(m.Email)] = m
	return nil
}

func (f *fakeMembers) GetMemberByEmail(_ context.Context, email string) (*model.Member, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.byEmail[strings.ToLower(email)]
	if !ok {
		return nil, store.ErrNotFound
	}
	return m, nil
}

func (f *fakeMembers) GetMemberByID(_ context.Context, id uuid.UUID) (*model.Member, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.members[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return m, nil
}

func (f *fakeMembers) ListMembers(_ context.Context) ([]model.Member, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]model.Member, 0, len(f.members))
	for _, m := range f.members {
		out = append(out, *m)
	}
	return out, nil
}

func (f *fakeMembers) CreateSession(_ context.Context, s *model.Session) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sessions[s.Token] = s
	return nil
}

func (f *fakeMembers) GetSessionByToken(_ context.Context, token string) (*model.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[token]
	if !ok {
		return nil, store.ErrNotFound
	}
	return s, nil
}

func (f *fakeMembers) DeleteSession(_ context.Context, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, token)
	return nil
}

func (f *fakeMembers) DeleteExpiredSessions(_ context.Context, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for t, s := range f.sessions {
		if s.ExpiresAt.Before(now) {
			delete(f.sessions, t)
		}
	}
	return nil
}

// --- BookStore fake ---

type fakeBooks struct {
	books map[uuid.UUID]*model.Book
	tags  map[uuid.UUID]map[string]bool
}

func newFakeBooks() *fakeBooks {
	return &fakeBooks{
		books: map[uuid.UUID]*model.Book{},
		tags:  map[uuid.UUID]map[string]bool{},
	}
}

func (f *fakeBooks) CreateBook(_ context.Context, b *model.Book) error {
	c := *b
	f.books[b.ID] = &c
	return nil
}

func (f *fakeBooks) GetBook(_ context.Context, id uuid.UUID) (*model.Book, error) {
	b, ok := f.books[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	c := *b
	for t := range f.tags[id] {
		c.Tags = append(c.Tags, t)
	}
	return &c, nil
}

func (f *fakeBooks) UpdateBookStatus(_ context.Context, id uuid.UUID, status string, finishedAt *time.Time) error {
	b, ok := f.books[id]
	if !ok {
		return store.ErrNotFound
	}
	b.Status = status
	b.FinishedAt = finishedAt
	return nil
}

func (f *fakeBooks) ListBooks(_ context.Context, status string) ([]model.Book, error) {
	var out []model.Book
	for _, b := range f.books {
		if status == "" || b.Status == status {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (f *fakeBooks) ListFinished(_ context.Context) ([]model.Book, error) {
	return f.ListBooks(context.Background(), model.StatusLido)
}

func (f *fakeBooks) DeleteBook(_ context.Context, id uuid.UUID) error {
	if _, ok := f.books[id]; !ok {
		return store.ErrNotFound
	}
	delete(f.books, id)
	return nil
}

func (f *fakeBooks) AddTag(_ context.Context, bookID uuid.UUID, tag string) error {
	if _, ok := f.tags[bookID]; !ok {
		f.tags[bookID] = map[string]bool{}
	}
	f.tags[bookID][tag] = true
	return nil
}

func (f *fakeBooks) RemoveTag(_ context.Context, bookID uuid.UUID, tag string) error {
	delete(f.tags[bookID], tag)
	return nil
}

func (f *fakeBooks) ListTags(_ context.Context, bookID uuid.UUID) ([]string, error) {
	var out []string
	for t := range f.tags[bookID] {
		out = append(out, t)
	}
	return out, nil
}

// --- VotingStore fake ---

type fakeVoting struct {
	rounds    map[uuid.UUID]*model.VoteRound
	roundBks  map[uuid.UUID][]uuid.UUID
	votes     map[uuid.UUID]map[uuid.UUID]*model.Vote // round -> member -> vote
	bookTitle map[uuid.UUID]string
}

func newFakeVoting() *fakeVoting {
	return &fakeVoting{
		rounds:    map[uuid.UUID]*model.VoteRound{},
		roundBks:  map[uuid.UUID][]uuid.UUID{},
		votes:     map[uuid.UUID]map[uuid.UUID]*model.Vote{},
		bookTitle: map[uuid.UUID]string{},
	}
}

func (f *fakeVoting) CreateRound(_ context.Context, r *model.VoteRound, bookIDs []uuid.UUID) error {
	f.rounds[r.ID] = r
	f.roundBks[r.ID] = append([]uuid.UUID{}, bookIDs...)
	f.votes[r.ID] = map[uuid.UUID]*model.Vote{}
	return nil
}

func (f *fakeVoting) GetOpenRound(_ context.Context) (*model.VoteRound, error) {
	for _, r := range f.rounds {
		if r.Status == model.RoundAberta {
			return r, nil
		}
	}
	return nil, store.ErrNotFound
}

func (f *fakeVoting) GetRound(_ context.Context, id uuid.UUID) (*model.VoteRound, error) {
	r, ok := f.rounds[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return r, nil
}

func (f *fakeVoting) ListRoundBooks(_ context.Context, roundID uuid.UUID) ([]model.Book, error) {
	ids := f.roundBks[roundID]
	out := make([]model.Book, 0, len(ids))
	for _, id := range ids {
		out = append(out, model.Book{ID: id, Title: f.bookTitle[id]})
	}
	return out, nil
}

func (f *fakeVoting) CastVote(_ context.Context, v *model.Vote) error {
	if _, ok := f.votes[v.RoundID]; !ok {
		f.votes[v.RoundID] = map[uuid.UUID]*model.Vote{}
	}
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	f.votes[v.RoundID][v.MemberID] = v
	return nil
}

func (f *fakeVoting) GetVoteByMember(_ context.Context, roundID, memberID uuid.UUID) (*model.Vote, error) {
	m, ok := f.votes[roundID]
	if !ok {
		return nil, store.ErrNotFound
	}
	v, ok := m[memberID]
	if !ok {
		return nil, store.ErrNotFound
	}
	return v, nil
}

func (f *fakeVoting) CountVotes(_ context.Context, roundID uuid.UUID) ([]model.VoteCount, error) {
	agg := map[uuid.UUID]int{}
	// Inicializa todos os livros da rodada com 0 para aparecerem na contagem.
	for _, bID := range f.roundBks[roundID] {
		agg[bID] = 0
	}
	for _, v := range f.votes[roundID] {
		agg[v.BookID]++
	}
	var out []model.VoteCount
	for bID, c := range agg {
		out = append(out, model.VoteCount{BookID: bID, Title: f.bookTitle[bID], Count: c})
	}
	// Ordena: maior count primeiro; empates por título.
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Count > out[i].Count ||
				(out[j].Count == out[i].Count && out[j].Title < out[i].Title) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}

func (f *fakeVoting) CloseRound(_ context.Context, id uuid.UUID, winner *uuid.UUID, closedAt time.Time) error {
	r, ok := f.rounds[id]
	if !ok {
		return store.ErrNotFound
	}
	r.Status = model.RoundEncerrada
	r.ClosedAt = &closedAt
	r.WinnerBookID = winner
	return nil
}

// --- ReviewStore fake ---

type fakeReviews struct {
	byBookMember map[string]*model.Review
}

func newFakeReviews() *fakeReviews {
	return &fakeReviews{byBookMember: map[string]*model.Review{}}
}

func reviewKey(bookID, memberID uuid.UUID) string {
	return bookID.String() + ":" + memberID.String()
}

func (f *fakeReviews) UpsertReview(_ context.Context, r *model.Review) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	c := *r
	f.byBookMember[reviewKey(r.BookID, r.MemberID)] = &c
	return nil
}

func (f *fakeReviews) GetReview(_ context.Context, bookID, memberID uuid.UUID) (*model.Review, error) {
	r, ok := f.byBookMember[reviewKey(bookID, memberID)]
	if !ok {
		return nil, store.ErrNotFound
	}
	return r, nil
}

func (f *fakeReviews) ListReviewsByBook(_ context.Context, bookID uuid.UUID) ([]model.Review, error) {
	var out []model.Review
	for _, r := range f.byBookMember {
		if r.BookID == bookID {
			out = append(out, *r)
		}
	}
	return out, nil
}

// --- MeetingStore fake ---

type fakeMeetings struct {
	meetings    map[uuid.UUID]*model.Meeting
	attendances map[uuid.UUID]map[uuid.UUID]*model.Attendance
	agenda      map[uuid.UUID][]model.AgendaItem
	members     map[uuid.UUID]*model.Member // para lembretes
}

func newFakeMeetings() *fakeMeetings {
	return &fakeMeetings{
		meetings:    map[uuid.UUID]*model.Meeting{},
		attendances: map[uuid.UUID]map[uuid.UUID]*model.Attendance{},
		agenda:      map[uuid.UUID][]model.AgendaItem{},
		members:     map[uuid.UUID]*model.Member{},
	}
}

func (f *fakeMeetings) CreateMeeting(_ context.Context, m *model.Meeting) error {
	c := *m
	f.meetings[m.ID] = &c
	return nil
}

func (f *fakeMeetings) GetMeeting(_ context.Context, id uuid.UUID) (*model.Meeting, error) {
	m, ok := f.meetings[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return m, nil
}

func (f *fakeMeetings) ListUpcoming(_ context.Context, from time.Time) ([]model.Meeting, error) {
	var out []model.Meeting
	for _, m := range f.meetings {
		if !m.Datetime.Before(from) {
			out = append(out, *m)
		}
	}
	return out, nil
}

func (f *fakeMeetings) ListDueForReminder(_ context.Context, now time.Time, horizon time.Duration) ([]model.Meeting, error) {
	end := now.Add(horizon)
	var out []model.Meeting
	for _, m := range f.meetings {
		if m.RemindedAt == nil && !m.Datetime.Before(now) && !m.Datetime.After(end) {
			out = append(out, *m)
		}
	}
	return out, nil
}

func (f *fakeMeetings) MarkReminded(_ context.Context, id uuid.UUID, at time.Time) error {
	m, ok := f.meetings[id]
	if !ok {
		return store.ErrNotFound
	}
	t := at
	m.RemindedAt = &t
	return nil
}

func (f *fakeMeetings) SetAttendance(_ context.Context, a *model.Attendance) error {
	if _, ok := f.attendances[a.MeetingID]; !ok {
		f.attendances[a.MeetingID] = map[uuid.UUID]*model.Attendance{}
	}
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	c := *a
	f.attendances[a.MeetingID][a.MemberID] = &c
	return nil
}

func (f *fakeMeetings) ListAttendances(_ context.Context, meetingID uuid.UUID) ([]model.Attendance, error) {
	m := f.attendances[meetingID]
	out := make([]model.Attendance, 0, len(m))
	for _, a := range m {
		c := *a
		if mb, ok := f.members[a.MemberID]; ok {
			c.Name = mb.Name
		}
		out = append(out, c)
	}
	return out, nil
}

func (f *fakeMeetings) ListAttendeesForReminder(_ context.Context, meetingID uuid.UUID) ([]model.Member, error) {
	var out []model.Member
	for memberID, a := range f.attendances[meetingID] {
		if a.Status == model.PresencaConfirmado || a.Status == model.PresencaTalvez {
			if mb, ok := f.members[memberID]; ok {
				out = append(out, *mb)
			}
		}
	}
	return out, nil
}

func (f *fakeMeetings) AddAgendaItem(_ context.Context, it *model.AgendaItem) error {
	if it.ID == uuid.Nil {
		it.ID = uuid.New()
	}
	f.agenda[it.MeetingID] = append(f.agenda[it.MeetingID], *it)
	return nil
}

func (f *fakeMeetings) ListAgenda(_ context.Context, meetingID uuid.UUID) ([]model.AgendaItem, error) {
	return append([]model.AgendaItem{}, f.agenda[meetingID]...), nil
}

// --- Email mock ---

type fakeMailer struct {
	Calls []struct {
		To      []string
		Subject string
		Body    string
	}
	Err error
}

func (f *fakeMailer) Send(to []string, subject, body string) error {
	if f.Err != nil {
		return f.Err
	}
	f.Calls = append(f.Calls, struct {
		To      []string
		Subject string
		Body    string
	}{to, subject, body})
	return nil
}
