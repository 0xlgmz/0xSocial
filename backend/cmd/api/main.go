package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/mailer"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/media"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/monetization"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool := postgresServer(context.Background())
	defer pool.Close()

	sender := configureMailer()
	mediaStore := configureMediaStore(context.Background())
	monetizationConfig, err := monetization.ConfigFromEnv()
	if err != nil {
		log.Fatalf("invalid monetization configuration: %v", err)
	}

	r := internal.NewRouter(pool, sender, mediaStore, monetizationConfig)

	server := &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: r,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func configureMailer() mailer.Sender {
	configuration, configured, err := mailer.ResendConfigFromEnv()
	if err != nil {
		log.Fatalf("invalid Resend configuration: %v", err)
	}
	if !configured {
		log.Print("Resend is not configured; emails will be written to the development log")
		return mailer.NewLogSender()
	}

	sender, err := mailer.NewResendSender(configuration)
	if err != nil {
		log.Fatalf("configure Resend: %v", err)
	}

	return sender
}

func configureMediaStore(ctx context.Context) media.Store {
	configuration, configured, err := media.R2ConfigFromEnv()
	if err != nil {
		log.Fatalf("invalid R2 configuration: %v", err)
	}
	if !configured {
		log.Print("R2 media storage is not configured; image uploads are disabled")
		return nil
	}

	store, err := media.NewR2Store(ctx, configuration)
	if err != nil {
		log.Fatalf("configure R2 media storage: %v", err)
	}

	return store
}

func postgresServer(ctx context.Context) *pgxpool.Pool {
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_DB"),
	)
	pool, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	return pool
}
