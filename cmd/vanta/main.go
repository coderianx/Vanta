package main

import (
	"log"

	"github.com/coderianx/Vanta/internal/handlers/auth"
	"github.com/coderianx/Vanta/internal/handlers/follow"
	"github.com/coderianx/Vanta/internal/handlers/pulse"
	"github.com/coderianx/Vanta/internal/handlers/users"
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

		r.Get("/users/{username}", users.GetUserProfileHandler)
		r.Get("/users/{username}/pulses", pulse.GetUserPulses)

		r.With(
			middlewares.AuthMiddleware,
		).Post("/users/{username}/follow", follow.FollowHandler)

		r.With(
			middlewares.AuthMiddleware,
		).Delete("/users/{username}/follow", follow.UnfollowHandler)

		r.With(
			middlewares.AuthMiddleware,
		).Get("/users/{username}/follow-status", follow.GetFollowStatusHandler)

		r.Get("/pulses/{id}", pulse.GetPulseByID)

		r.With(
			middlewares.AuthMiddleware,
		).Post("/pulses", pulse.CreateNewPulse)

		r.With(
			middlewares.AuthMiddleware,
		).Delete("/pulses/{id}", pulse.DeleteMyPulse)
	})
}
