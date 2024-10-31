package main

import (
	"context"
	"fmt"
	strLog "log"
	"os/signal"
	"syscall"

	"github.com/allegro/bigcache/v3"
	"go.uber.org/zap"

	"github.com/sosnovski/test_work_2/internal/config"
	"github.com/sosnovski/test_work_2/internal/handler"
	"github.com/sosnovski/test_work_2/internal/proto"
	"github.com/sosnovski/test_work_2/internal/server"
)

const (
	QuoteResourceID proto.ResourceIDType = 0
	TimeResourceID  proto.ResourceIDType = 1
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		strLog.Fatal(fmt.Errorf("create zap logger: %w", err))
	}

	cfg, err := config.InitServer(ctx, "")
	if err != nil {
		logger.Fatal("init config", zap.Error(err))
	}

	quotes := []string{
		"Yesterday I was clever, so I wanted to change the world. Today I am wise, so I am changing myself.",
		"I know not with what weapons World War III will be fought, " +
			"but World War IV will be fought with sticks and stones.",
		"Don't Gain The World & Lose Your Soul, Wisdom Is Better Than Silver Or Gold.",
		"I’m not in this world to live up to your expectations and you’re not in this world to live up to mine.",
	}

	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(cfg.ChallengeCacheTTL))
	if err != nil {
		logger.Fatal("init cache", zap.Error(err))
	}

	srv := server.NewServer(
		logger,
		cfg.ReadTimeout,
		cfg.WriteTimeout,
		cfg.PowDifficulty,
		cfg.ChallengeTimeout,
		cfg.ChallengeRandBytesCount,
		[]byte(cfg.Secret),
		cache,
	)

	h := handler.NewQuoteHandler(quotes)

	err = srv.RegisterHandlers(
		QuoteResourceID, h.RandomQuote,
		TimeResourceID, h.CurrentTime,
	)
	if err != nil {
		logger.Fatal("register handler", zap.Error(err))
	}

	stop, err := srv.Listen(ctx, cfg.ListenAddress)
	if err != nil {
		logger.Fatal("start listen", zap.Error(err))
	}

	<-ctx.Done()

	shutdownCtx := context.Background()
	if cfg.ShutdownTimeout > 0 {
		ctx, cancel := context.WithTimeout(shutdownCtx, cfg.ShutdownTimeout)
		defer cancel()

		shutdownCtx = ctx
	}

	if err := stop(shutdownCtx); err != nil {
		logger.Error("stop server", zap.Error(err))
	}
}
