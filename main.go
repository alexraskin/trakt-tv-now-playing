package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexraskin/trakt-tv-now-playing/server"
)

var (
	version   = "unknown"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {

	port := flag.String("port", "4000", "port to listen on")
	flag.Parse()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	slog.Info("Starting trakt-tv-now-playing...", slog.Any("version", version), slog.Any("commit", commit), slog.Any("buildTime", buildTime))

	server := server.NewServer(
		server.GetTraktAPIKey(),
		server.FormatBuildVersion(version, commit, buildTime),
		*port,
		httpClient,
	)

	go server.Start()
	defer server.Close()

	slog.Info("started web server", slog.Any("listen_addr", *port))
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si
	slog.Info("shutting down web server")
}
