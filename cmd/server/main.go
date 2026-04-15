// Entrypoint do Clube do Livro. Conecta no Postgres, roda migrações,
// semeia o admin, sobe o servidor HTTP e agenda o envio de lembretes.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clube-do-livro/app/internal/handler"
	"github.com/clube-do-livro/app/internal/service"
	"github.com/clube-do-livro/app/internal/store"
	"github.com/clube-do-livro/app/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "roda as migrações e encerra")
	flag.Parse()

	dsn := env("DATABASE_URL", "")
	if dsn == "" {
		log.Fatal("DATABASE_URL não configurado")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("falha ao conectar no Postgres: %v", err)
	}
	defer pool.Close()

	// Aguarda o banco com pequeno retry.
	if err := waitForDB(ctx, pool); err != nil {
		log.Fatalf("banco indisponível: %v", err)
	}

	// Migrações (embutidas no binário).
	if err := store.Migrate(ctx, pool, migrations.FS); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrações em dia")

	if *migrateOnly {
		return
	}

	// Stores + serviços.
	pg := store.New(pool)
	auth := service.NewAuth(pg)
	books := service.NewBook(pg)
	search := service.NewBookSearch(env("GOOGLE_BOOKS_API_KEY", ""))
	voting := service.NewVoting(pg)
	reviews := service.NewReview(pg, pg)
	mailer := &service.SMTPMailer{
		Host: os.Getenv("SMTP_HOST"),
		Port: env("SMTP_PORT", "587"),
		User: os.Getenv("SMTP_USER"),
		Pass: os.Getenv("SMTP_PASS"),
		From: env("SMTP_FROM", "clube@exemplo.com"),
	}
	meetings := service.NewMeeting(pg, mailer)

	// Seed do admin inicial, se não houver.
	if err := seedAdmin(ctx, auth); err != nil {
		log.Printf("aviso: seed do admin falhou: %v", err)
	}

	// Roteador HTTP.
	r := handler.New(handler.Deps{
		Auth:     auth,
		Books:    books,
		Search:   search,
		Voting:   voting,
		Reviews:  reviews,
		Meetings: meetings,
	})

	port := env("PORT", "8080")
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Goroutine dedicada aos lembretes (verifica a cada 15 min).
	stop := make(chan struct{})
	go reminderLoop(meetings, stop)

	// Graceful shutdown.
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		log.Println("encerrando...")
		close(stop)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	log.Printf("Clube do Livro escutando em :%s", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}

func reminderLoop(m *service.MeetingService, stop <-chan struct{}) {
	t := time.NewTicker(15 * time.Minute)
	defer t.Stop()
	// Roda uma vez imediatamente também.
	if n, err := m.SendDueReminders(context.Background()); err != nil {
		log.Printf("reminders: %v", err)
	} else if n > 0 {
		log.Printf("%d lembrete(s) enviado(s)", n)
	}
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			if n, err := m.SendDueReminders(context.Background()); err != nil {
				log.Printf("reminders: %v", err)
			} else if n > 0 {
				log.Printf("%d lembrete(s) enviado(s)", n)
			}
		}
	}
}

func seedAdmin(ctx context.Context, a *service.AuthService) error {
	email := env("ADMIN_EMAIL", "admin@clube.local")
	pass := env("ADMIN_PASSWORD", "admin123")
	name := env("ADMIN_NAME", "Administrador")
	if _, err := a.Members.GetMemberByEmail(ctx, email); err == nil {
		return nil // já existe
	}
	_, err := a.CreateMember(ctx, name, email, pass, true)
	if err != nil {
		return err
	}
	log.Printf("admin inicial criado: %s", email)
	return nil
}

func waitForDB(ctx context.Context, pool *pgxpool.Pool) error {
	deadline := time.Now().Add(30 * time.Second)
	for {
		if err := pool.Ping(ctx); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("timeout esperando o banco")
		}
		time.Sleep(1 * time.Second)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
