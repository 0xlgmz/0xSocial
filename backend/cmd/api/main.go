package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func init() {
	// Loading the environment variables from '.env' file.
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("unable to load .env file: %e", err)
	}

}

func main() {
	apiServer(os.Getenv("GO_ADDR"))
}

func apiServer(addr string) {
	pool := postgresServer(context.Background())
	r := internal.NewRouter()
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	log.Printf("Starting server on %s", addr)
	server.ListenAndServe()
	defer pool.Close()
}
func postgresServer(ctx context.Context) *pgxpool.Pool {
	databaseURL := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)
	log.Println(databaseURL)
	pool, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	return pool
}
