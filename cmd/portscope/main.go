package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/sunsh1ny/portscope/internal/app"
)

func main() {
	ctx := context.Background()

	if err := app.Run(ctx, os.Args[1:]); err != nil {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		logger.Error("application failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
