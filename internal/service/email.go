package service

import (
	"fmt"
	"log"
	"net/smtp"
)

// SMTPMailer é o EmailSender de produção. Se Host estiver vazio,
// apenas registra no log — o que mantém a aplicação funcional sem SMTP.
type SMTPMailer struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

func (s *SMTPMailer) Send(to []string, subject, body string) error {
	if s.Host == "" {
		log.Printf("[email mock] para=%v assunto=%q corpo=%q", to, subject, body)
		return nil
	}
	addr := s.Host + ":" + s.Port
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.From, joinCSV(to), subject, body))
	var auth smtp.Auth
	if s.User != "" {
		auth = smtp.PlainAuth("", s.User, s.Pass, s.Host)
	}
	return smtp.SendMail(addr, auth, s.From, to, msg)
}

func joinCSV(xs []string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += ", "
		}
		out += x
	}
	return out
}
