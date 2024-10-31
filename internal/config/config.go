package config

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vrischmann/envconfig"
)

type ClientConfig struct {
	ServerAddress           string        `envconfig:"default=:6543"`
	ComputeChallengeTimeout time.Duration `envconfig:"default=300ms"`
	RequestTimeout          time.Duration `envconfig:"default=1s"`
}

type ServerConfig struct {
	ListenAddress           string        `envconfig:"default=:6543"`
	Secret                  string        `envconfig:"default=some_default_secret" validate:"min=16"`
	ShutdownTimeout         time.Duration `envconfig:"default=5s"`
	ReadTimeout             time.Duration `envconfig:"default=500ms"`
	WriteTimeout            time.Duration `envconfig:"default=500ms"`
	PowDifficulty           uint8         `envconfig:"default=18"                  validate:"min=0,max=255"`
	ChallengeTimeout        time.Duration `envconfig:"default=400ms"`
	ChallengeCacheTTL       time.Duration `envconfig:"default=500ms"`
	ChallengeRandBytesCount int           `envconfig:"default=8"`
}

func InitServer(ctx context.Context, prefix string) (*ServerConfig, error) {
	cfg := &ServerConfig{}
	if err := envconfig.InitWithPrefix(cfg, prefix); err != nil {
		return nil, fmt.Errorf("failed to init config: %w", err)
	}

	if err := validator.New().StructCtx(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func InitClient(ctx context.Context, prefix string) (*ClientConfig, error) {
	cfg := &ClientConfig{}
	if err := envconfig.InitWithPrefix(cfg, prefix); err != nil {
		return nil, fmt.Errorf("failed to init config: %w", err)
	}

	if err := validator.New().StructCtx(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}
