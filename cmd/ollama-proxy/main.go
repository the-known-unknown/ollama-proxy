package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asimmittal/ollama-proxy/internal/auth"
	"github.com/asimmittal/ollama-proxy/internal/config"
	"github.com/asimmittal/ollama-proxy/internal/proxy"
	"github.com/asimmittal/ollama-proxy/internal/upstream"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(args []string) error {
	cfg, err := config.Parse(args)
	if err != nil {
		return err
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	if !cfg.SecurityEnabled() && !cfg.Insecure {
		ok, err := config.ConfirmInsecure(os.Stdin, os.Stdout)
		if err != nil {
			return err
		}
		if !ok {
			logger.Println("aborted: no API key provided")
			return nil
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client := upstream.NewClient(cfg.Host)
	platform := upstream.NewOllamaPlatform(client)

	if !platform.Running(ctx) {
		logger.Printf("platform %q not responding at %s; attempting to start it", cfg.Platform, cfg.Host)
		started, err := platform.EnsureRunning(ctx, 15*time.Second, 500*time.Millisecond)
		if err != nil {
			return fmt.Errorf("could not start platform: %w", err)
		}
		if started {
			logger.Println("platform started successfully")
		}
	}

	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("upstream host is not alive: %w", err)
	}

	if cfg.Model != "" {
		ok, err := client.HasModel(ctx, cfg.Model)
		if err != nil {
			return fmt.Errorf("verify model: %w", err)
		}
		if !ok {
			return fmt.Errorf("model %q not found at %s/v1/models", cfg.Model, cfg.Host)
		}
		logger.Printf("model %q verified", cfg.Model)
	} else {
		models, err := client.ListModels(ctx)
		if err != nil {
			return fmt.Errorf("list models: %w", err)
		}
		logger.Printf("no model specified; %d models available upstream:", len(models))
		for _, m := range models {
			logger.Printf("  - %s", m.ID)
		}
	}

	targetURL, err := cfg.HostURL()
	if err != nil {
		return err
	}

	handler := auth.Middleware(cfg.APIKey)(proxy.New(targetURL))

	if cfg.SecurityEnabled() {
		logger.Println("API key security: enabled")
	} else {
		logger.Println("API key security: DISABLED")
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Printf("listening on %s -> %s", addr, cfg.Host)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Println("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
