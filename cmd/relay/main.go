package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tomrford/monitor-cli/internal/relay"
)

const (
	envListenAddr      = "RELAY_LISTEN_ADDR"
	envForwardURL      = "RELAY_FORWARD_URL"
	envForwardToken    = "RELAY_FORWARD_TOKEN"
	envWebhookSecret   = "PARALLEL_WEBHOOK_SECRET"
	envMetadataToken   = "RELAY_METADATA_TOKEN"
	envAllowedMonitors = "RELAY_ALLOW_MONITOR_IDS"
	envMaxBodyBytes    = "RELAY_MAX_BODY_BYTES"
	envReplayWindowSec = "RELAY_REPLAY_WINDOW_SECONDS"
	envRPS             = "RELAY_RPS"
	envBurst           = "RELAY_BURST"
)

func main() {
	if err := run(); err != nil {
		log.Printf("error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := relay.LoadConfigFromEnv(relay.Env{
		ListenAddr:      envListenAddr,
		ForwardURL:      envForwardURL,
		ForwardToken:    envForwardToken,
		WebhookSecret:   envWebhookSecret,
		MetadataToken:   envMetadataToken,
		AllowedMonitors: envAllowedMonitors,
		MaxBodyBytes:    envMaxBodyBytes,
		ReplayWindowSec: envReplayWindowSec,
		RPS:             envRPS,
		Burst:           envBurst,
	})
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           relay.NewServer(cfg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening addr=%s forward_url=%s", cfg.ListenAddr, cfg.ForwardURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-stop:
		log.Printf("shutdown")
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
