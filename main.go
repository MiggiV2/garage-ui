package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/MiggiV2/garage-ui/internal/config"
	"github.com/MiggiV2/garage-ui/internal/server"
)

//go:embed all:web/dist
var dist embed.FS

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	assets, err := fs.Sub(dist, "web/dist")
	if err != nil {
		slog.Error("embedded assets missing", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           server.New(cfg, assets),
		ReadHeaderTimeout: 10 * time.Second,
	}
	slog.Info("garage-ui listening", "addr", cfg.ListenAddr, "garage", cfg.AdminURL, "basicAuth", cfg.BasicAuthEnabled())
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
