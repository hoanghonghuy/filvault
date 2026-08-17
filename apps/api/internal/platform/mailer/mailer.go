package mailer

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"sync"

	"filnest/internal/platform/config"
)

type Message struct {
	To      string
	Subject string
	Body    string
}

type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

// New picks console when SMTP is not configured; otherwise SMTP.
// FILNEST_MAILER=console forces console even if SMTP host is set.
func New(cfg config.Config) Mailer {
	if strings.EqualFold(cfg.Mailer, "console") || cfg.SMTPHost == "" {
		return Console{}
	}
	return SMTP{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUsername,
		Password: cfg.SMTPPassword,
		From:     cfg.SMTPFrom,
	}
}

// Console logs the message body (used when no SMTP is configured).
type Console struct{}

func (Console) Send(_ context.Context, msg Message) error {
	slog.Info("mail", "to", msg.To, "subject", msg.Subject, "body", msg.Body)
	return nil
}

// SMTP sends mail via net/smtp.
type SMTP struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func (s SMTP) Send(_ context.Context, msg Message) error {
	from := s.From
	if from == "" {
		from = s.Username
	}
	addr := s.Host + ":" + s.Port
	body := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", msg.To, msg.Subject, msg.Body)
	var auth smtp.Auth
	if s.Username != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{msg.To}, []byte(body))
}

// Memory captures messages in-process; used by tests.
type Memory struct {
	mu    sync.Mutex
	codes map[string]string
}

func NewMemory() *Memory {
	return &Memory{codes: map[string]string{}}
}

func (m *Memory) Send(_ context.Context, msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[msg.To] = msg.Body
	return nil
}

func (m *Memory) LastCode(to string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.codes[to]
}
