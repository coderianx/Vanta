package main

import (
	"log"

	"github.com/coderianx/Vanta/internal/handlers/auth"
	"github.com/coderianx/Vanta/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()

	router.Route("/api/v1", func(r chi.Router) {

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.RegisterHandler)
			r.Post("/login", auth.LoginHandler)
			r.Post("/refresh", auth.RefreshHandler)
		})

		r.With(
			middlewares.AuthMiddleware,
		).Get("/me", auth.GetMe)
	})
}
