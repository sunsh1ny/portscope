package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/sunsh1ny/portscope/internal/config"
	"github.com/sunsh1ny/portscope/internal/scanner"
)

func Run(parent context.Context, args []string) error {
	cfg, err := config.Parse(args)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	logger := newLogger(cfg.JSONLogs)

	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info(
		"starting scan",
		slog.String("host", cfg.Host),
		slog.Int("start_port", cfg.StartPort),
		slog.Int("end_port", cfg.EndPort),
		slog.Int("workers", cfg.Workers),
		slog.Duration("timeout", cfg.Timeout),
	)

	startedAt := time.Now()

	s := scanner.New(scanner.Config{
		Host:    cfg.Host,
		Start:   cfg.StartPort,
		End:     cfg.EndPort,
		Workers: cfg.Workers,
		Timeout: cfg.Timeout,
	})

	results, err := s.Scan(ctx)
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("scan ports: %w", err)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	printResults(results)

	logger.Info(
		"scan finished",
		slog.Int("open_ports", len(results)),
		slog.Duration("duration", time.Since(startedAt)),
	)

	if ctx.Err() != nil {
		logger.Warn("scan interrupted", slog.String("reason", ctx.Err().Error()))
	}

	return nil
}

func newLogger(json bool) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if json {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func printResults(results []scanner.Result) {
	if len(results) == 0 {
		fmt.Println("No open TCP ports found")
		return
	}

	fmt.Println("Open TCP ports:")
	for _, result := range results {
		fmt.Printf("- %d\n", result.Port)
	}
}
