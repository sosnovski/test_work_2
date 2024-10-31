package main

import (
	"context"
	"fmt"
	strLog "log"
	"os/signal"
	"syscall"

	"github.com/sosnovski/test_work_2/client"
	"github.com/sosnovski/test_work_2/internal/config"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.PanicLevel))
	if err != nil {
		strLog.Fatal(fmt.Errorf("create zap logger: %w", err))
	}

	cfg, err := config.InitClient(ctx, "")
	if err != nil {
		logger.Fatal("init config", zap.Error(err))
	}

	cli := client.New(cfg.ServerAddress, cfg.ComputeChallengeTimeout, cfg.RequestTimeout)

	defer logger.Info("client stopped")

	logger.Info("client started", zap.String("server_address", cfg.ServerAddress))

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		quote, err := cli.Quote(ctx)
		if err != nil {
			logger.Error("get quote", zap.Error(err))

			continue
		}

		logger.Info("quote received", zap.Any("quote", quote))
	}
}
