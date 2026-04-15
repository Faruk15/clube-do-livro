package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/clube-do-livro/app/internal/model"
	"github.com/clube-do-livro/app/internal/store"
	"github.com/google/uuid"
)

// EmailSender é injetado para permitir mock nos testes.
type EmailSender interface {
	Send(to []string, subject, body string) error
}

// MeetingService — encontros, presença, pauta e lembretes.
type MeetingService struct {
	Meetings store.MeetingStore
	Mailer   EmailSender // pode ser nil: apenas não envia
	Now      func() time.Time
}

func NewMeeting(m store.MeetingStore, mailer EmailSender) *MeetingService {
	return &MeetingService{Meetings: m, Mailer: mailer, Now: time.Now}
}

// Create agenda um encontro. Valida presença de título e data futura.
func (m *MeetingService) Create(ctx context.Context, meet *model.Meeting) error {
	if meet.Title == "" {
		return errors.New("título obrigatório")
	}
	if meet.Datetime.IsZero() {
		return errors.New("data/hora obrigatória")
	}
	meet.ID = uuid.New()
	meet.CreatedAt = m.Now()
	return m.Meetings.CreateMeeting(ctx, meet)
}

func (m *MeetingService) Upcoming(ctx context.Context) ([]model.Meeting, error) {
	return m.Meetings.ListUpcoming(ctx, m.Now())
}

func (m *MeetingService) Get(ctx context.Context, id uuid.UUID) (*model.Meeting, error) {
	return m.Meetings.GetMeeting(ctx, id)
}

// SetAttendance grava (ou atualiza) a resposta do membro.
func (m *MeetingService) SetAttendance(ctx context.Context, meetingID, memberID uuid.UUID, status string) error {
	switch status {
	case model.PresencaConfirmado, model.PresencaNaoVou, model.PresencaTalvez:
	default:
		return errors.New("status de presença inválido")
	}
	return m.Meetings.SetAttendance(ctx, &model.Attendance{
		MeetingID: meetingID, MemberID: memberID, Status: status,
	})
}

func (m *MeetingService) Attendances(ctx context.Context, meetingID uuid.UUID) ([]model.Attendance, error) {
	return m.Meetings.ListAttendances(ctx, meetingID)
}

// AddAgendaItem insere um tópico de pauta.
func (m *MeetingService) AddAgendaItem(ctx context.Context, meetingID, memberID uuid.UUID, content string) error {
	if content == "" {
		return errors.New("conteúdo vazio")
	}
	return m.Meetings.AddAgendaItem(ctx, &model.AgendaItem{
		MeetingID: meetingID, MemberID: memberID, Content: content,
	})
}

func (m *MeetingService) Agenda(ctx context.Context, meetingID uuid.UUID) ([]model.AgendaItem, error) {
	return m.Meetings.ListAgenda(ctx, meetingID)
}

// SendDueReminders procura encontros em até 24h que ainda não foram
// lembrados e envia e-mails aos confirmados/talvez. Seguro p/ rodar várias vezes.
func (m *MeetingService) SendDueReminders(ctx context.Context) (int, error) {
	if m.Mailer == nil {
		return 0, nil
	}
	now := m.Now()
	meets, err := m.Meetings.ListDueForReminder(ctx, now, 24*time.Hour)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, meet := range meets {
		members, err := m.Meetings.ListAttendeesForReminder(ctx, meet.ID)
		if err != nil {
			return sent, err
		}
		if len(members) == 0 {
			_ = m.Meetings.MarkReminded(ctx, meet.ID, now)
			continue
		}
		emails := make([]string, 0, len(members))
		for _, mb := range members {
			emails = append(emails, mb.Email)
		}
		subject := fmt.Sprintf("Lembrete: encontro do clube — %s", meet.Title)
		body := fmt.Sprintf(
			"Olá!\n\nLembrete: o encontro \"%s\" acontece em %s.\nLocal: %s\n\nAté lá!\nClube do Livro",
			meet.Title, meet.Datetime.Format("02/01/2006 15:04"), nvlLocal(meet.Location))
		if err := m.Mailer.Send(emails, subject, body); err != nil {
			return sent, err
		}
		if err := m.Meetings.MarkReminded(ctx, meet.ID, now); err != nil {
			return sent, err
		}
		sent++
	}
	return sent, nil
}

func nvlLocal(s string) string {
	if s == "" {
		return "(a combinar)"
	}
	return s
}
