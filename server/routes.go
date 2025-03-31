package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

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

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Get("/{username}", s.handleCheckWatching)

	r.NotFound(s.notFound)

	return r
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
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

	req, err := http.NewRequest("GET", "https://api.trakt.tv/users/"+username+"/watching", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("Error creating request", slog.Any("err", err))
		return
	}
	req.Header.Add("trakt-api-key", s.traktApiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("trakt-api-version", "2")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("Error sending request", slog.Any("err", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("Error reading response body", slog.Any("err", err))
		return
	}

	if resp.StatusCode == http.StatusNoContent {
		if format == "shields.io" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"schemaVersion": 1, "label": "Currently Watching", "message": "Nothing", "color": "red"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"watching": false}`))
		}
		return
	}

	var watchingResponse TraktWatchingResponse
	if err := json.Unmarshal(body, &watchingResponse); err != nil {
		slog.Error("Error unmarshalling response body", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if format == "shields.io" {
		message := "Nothing"
		if watchingResponse.Movie != nil {
			message = watchingResponse.Movie.Title
		} else if watchingResponse.Show != nil {
			if watchingResponse.Episode != nil {
				message = fmt.Sprintf("%s S%02dE%02d",
					watchingResponse.Show.Title,
					watchingResponse.Episode.Season,
					watchingResponse.Episode.Number)
			} else {
				message = watchingResponse.Show.Title
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"schemaVersion": 1, "label": "Currently Watching", "message": "` + message + `", "color": "green"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
