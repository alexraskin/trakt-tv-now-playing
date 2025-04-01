package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Origin, Content-Type, Accept"},
		AllowedMethods: []string{"GET", "OPTIONS"},
	}))

	r.Use(httprate.Limit(
		100,
		time.Minute,
		httprate.WithLimitHandler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
			}),
		),
	))

	r.Get("/", s.index)
	r.Get("/version", s.serverVersion)
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Get("/{username}", s.handleCheckWatching)

	r.NotFound(s.notFound)

	return r
}

func (s *Server) serverVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s.version))
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trakt TV Now Playing"))
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Found", http.StatusNotFound)
}

func (s *Server) handleCheckWatching(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	format := r.URL.Query().Get("format")

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	body, err := s.getTraktUser(r.Context(), username)
	if err != nil {
		switch err.Error() {
		case "user not found":
			http.Error(w, "User not found", http.StatusNotFound)
		default:
			slog.Error("Error fetching Trakt user data", slog.Any("err", err))
			http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		}
		return
	}

	if body == (TraktWatchingResponse{}) {
		if format == "shields.io" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"schemaVersion": 1, "label": "Currently Watching", "message": "Nothing", "color": "red"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"watching": false}`))
		}
		return
	}

	if format == "shields.io" {
		message := "Nothing"
		if body.Movie != nil {
			message = body.Movie.Title
		} else if body.Show != nil {
			if body.Episode != nil {
				message = fmt.Sprintf("%s S%02dE%02d",
					body.Show.Title,
					body.Episode.Season,
					body.Episode.Number)
			} else {
				message = body.Show.Title
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"schemaVersion": 1, "label": "Currently Watching", "message": "` + message + `", "color": "green"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
}
