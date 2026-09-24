package httpapi

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetProfile(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
func UpdateProfile(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
