package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/aohta/home-portal/internal/httpapi"
	"github.com/aohta/home-portal/internal/webui"
)

var Version = "dev"

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(stdout, "Usage: home-portal serve [--listen 127.0.0.1:8082] [--homehub-url http://127.0.0.1:8081] [--assets directory]\n       home-portal version")
		return 0
	}
	if args[0] == "version" && len(args) == 1 {
		fmt.Fprintln(stdout, Version)
		return 0
	}
	if args[0] != "serve" {
		fmt.Fprintln(stderr, "unknown command: use serve or version")
		return 2
	}
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.SetOutput(stderr)
	listen := flags.String("listen", "127.0.0.1:8082", "HTTP listen address")
	source := flags.String("homehub-url", "http://127.0.0.1:8081", "Homehub API base URL")
	assets := flags.String("assets", "", "external frontend directory (development only)")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *listen == "" {
		fmt.Fprintln(stderr, "a non-empty --listen is required; positional arguments are not accepted")
		return 2
	}
	logger := slog.New(slog.NewJSONHandler(stderr, nil))
	web, err := webui.Handler(*assets)
	if err != nil {
		logger.Error("load frontend", "error", err)
		return 1
	}
	handler, err := httpapi.New(*source, web, logger)
	if err != nil {
		logger.Error("configure API", "error", err)
		return 2
	}
	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		logger.Error("listen", "error", err)
		return 1
	}
	server := &http.Server{
		Handler: handler, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	logger.Info("HTTP server listening", "address", listener.Addr().String(), "version", Version)
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serve HTTP", "error", err)
			return 1
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			logger.Error("shutdown HTTP", "error", err)
			return 1
		}
	}
	return 0
}
