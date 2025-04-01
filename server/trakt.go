package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

func (s *Server) getTraktUser(ctx context.Context, username string) (TraktWatchingResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.trakt.tv/users/"+username+"/watching", nil)
	if err != nil {
		return TraktWatchingResponse{}, err
	}
	req.Header.Add("trakt-api-key", s.traktApiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("trakt-api-version", "2")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return TraktWatchingResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return TraktWatchingResponse{}, errors.New("user not found")
	}

	if resp.StatusCode == http.StatusNoContent {
		return TraktWatchingResponse{}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TraktWatchingResponse{}, err
	}

	var watchingResponse TraktWatchingResponse
	if err := json.Unmarshal([]byte(body), &watchingResponse); err != nil {
		slog.Error("Error unmarshalling response body", slog.Any("err", err))
		return TraktWatchingResponse{}, err
	}

	return watchingResponse, nil
}
