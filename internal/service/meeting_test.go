package service

import (
	"context"
	"testing"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMeeting_LembretesEnviamApenasConfirmadosETalvez(t *testing.T) {
	ctx := context.Background()
	fm := newFakeMeetings()
	mailer := &fakeMailer{}
	svc := NewMeeting(fm, mailer)

	base := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)
	svc.Now = func() time.Time { return base }

	// Encontro daqui a 6h (dentro da janela de 24h)
	meet := &model.Meeting{Title: "Leitura", Datetime: base.Add(6 * time.Hour), Location: "Café"}
	require.NoError(t, svc.Create(ctx, meet))

	ana := &model.Member{ID: uuid.New(), Email: "ana@x", Name: "Ana"}
	bob := &model.Member{ID: uuid.New(), Email: "bob@x", Name: "Bob"}
	carol := &model.Member{ID: uuid.New(), Email: "carol@x", Name: "Carol"}
	fm.members[ana.ID] = ana
	fm.members[bob.ID] = bob
	fm.members[carol.ID] = carol

	require.NoError(t, svc.SetAttendance(ctx, meet.ID, ana.ID, model.PresencaConfirmado))
	require.NoError(t, svc.SetAttendance(ctx, meet.ID, bob.ID, model.PresencaNaoVou))
	require.NoError(t, svc.SetAttendance(ctx, meet.ID, carol.ID, model.PresencaTalvez))

	sent, err := svc.SendDueReminders(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, sent)
	require.Len(t, mailer.Calls, 1)
	require.ElementsMatch(t, []string{"ana@x", "carol@x"}, mailer.Calls[0].To)

	// idempotente: segunda execução não manda de novo
	sent, err = svc.SendDueReminders(ctx)
	require.NoError(t, err)
	require.Zero(t, sent)
}

func TestMeeting_LembreteSoNaJanelaDe24h(t *testing.T) {
	ctx := context.Background()
	fm := newFakeMeetings()
	mailer := &fakeMailer{}
	svc := NewMeeting(fm, mailer)
	base := time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC)
	svc.Now = func() time.Time { return base }

	// daqui a 3 dias — fora da janela
	meet := &model.Meeting{Title: "Futuro", Datetime: base.Add(72 * time.Hour)}
	require.NoError(t, svc.Create(ctx, meet))

	_, err := svc.SendDueReminders(ctx)
	require.NoError(t, err)
	require.Empty(t, mailer.Calls)
}

func TestMeeting_CreateValida(t *testing.T) {
	svc := NewMeeting(newFakeMeetings(), nil)
	err := svc.Create(context.Background(), &model.Meeting{})
	require.Error(t, err)
}

func TestMeeting_StatusPresencaInvalido(t *testing.T) {
	ctx := context.Background()
	fm := newFakeMeetings()
	svc := NewMeeting(fm, nil)
	err := svc.SetAttendance(ctx, uuid.New(), uuid.New(), "chumbado")
	require.Error(t, err)
}

func TestMeeting_SemMailerNaoFalha(t *testing.T) {
	ctx := context.Background()
	fm := newFakeMeetings()
	svc := NewMeeting(fm, nil)
	svc.Now = func() time.Time { return time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC) }
	meet := &model.Meeting{Title: "x", Datetime: svc.Now().Add(time.Hour)}
	require.NoError(t, svc.Create(ctx, meet))

	n, err := svc.SendDueReminders(ctx)
	require.NoError(t, err)
	require.Zero(t, n)
}
