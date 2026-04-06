// SPDX-FileCopyrightText: 2026-present Igor Kha.
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/napilab/discoverd/internal/app"
	"github.com/napilab/discoverd/internal/config"
	"github.com/napilab/discoverd/version"
	"github.com/spf13/pflag"
)

func main() {
	if hasVersionFlag(os.Args[1:]) {
		if err := printVersion(os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return
	}

	cfg, err := config.Parse(os.Args[1:], os.Getenv)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg, logger, os.Stdout); err != nil {
		logger.Println(err)
		os.Exit(1)
	}
}

func hasVersionFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--version" || arg == "-V" {
			return true
		}
	}

	return false
}

func printVersion(out io.Writer) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(version.Info()); err != nil {
		return fmt.Errorf("encode version info: %w", err)
	}

	return nil
}
