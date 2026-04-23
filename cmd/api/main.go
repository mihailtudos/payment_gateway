package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/docs"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/config"
	"github.com/cko-recruitment/payment-gateway-challenge-go/pkg/tel"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

//	@title			Payment Gateway Challenge Go
//	@description	Interview challenge for building a Payment Gateway - Go version
//	@host		localhost:8090
//	@BasePath	/
//
// @securityDefinitions.basic	BasicAuth
func main() {
	docs.SwaggerInfo.Version = version

	err := run()
	if err != nil {
		slog.Error("fatal API error", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustNewConfig(version)

	telemetry, err := tel.New(ctx, tel.Config{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		OTLPEndpoint:   cfg.OTLPEndpoint,
	})
	if err != nil {
		slog.Warn("OTel - running as no-op", slog.Any("err", err))
		telemetry = tel.NewNoopTelemetry()
	}

	defer func() {
		slog.Info("flushing telemetry...")
		shutCtx, shutCancel := context.WithTimeout(
			context.Background(),
			time.Duration(cfg.OTelShutdownTimeout)*time.Second,
		)
		defer shutCancel()
		_ = telemetry.Shutdown(shutCtx)
	}()
	defer slog.Info("stopping server...")

	slog.SetDefault(telemetry.Logger)
	slog.Info("build info",
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("buildAt", date),
	)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		slog.Error("sigterm/interrupt signal")
		cancel()
	}()

	a := api.New(cfg.BankURL, telemetry)
	return a.Run(ctx, cfg.HTTP)
}
