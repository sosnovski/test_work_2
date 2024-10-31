package pow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChallenge_VerifySign(t *testing.T) {
	t.Parallel()

	var (
		validSecret         = []byte("valid_secret")
		invalidSecret       = []byte("test_secret")
		rand                = []byte{0x01, 0x02, 0x03, 0x04}
		dif           uint8 = 1
		unix                = time.Now().Unix()
	)

	tests := []struct {
		name          string
		makeChallenge func() *Challenge
		wantErr       error
		args          []byte
	}{
		{
			name: "valid secret",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, dif, unix, validSecret)
			},
			args:    validSecret,
			wantErr: nil,
		},
		{
			name: "invalid secret",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, dif, unix, validSecret)
			},
			args:    invalidSecret,
			wantErr: ErrInvalidSignature,
		},
		{
			name: "updated rand",
			makeChallenge: func() *Challenge {
				c := NewChallenge(rand, dif, unix, validSecret)
				c.Rand = append(c.Rand, 0x01)

				return c
			},
			args:    validSecret,
			wantErr: ErrInvalidSignature,
		},
		{
			name: "updated difficulty",
			makeChallenge: func() *Challenge {
				c := NewChallenge(rand, dif, unix, validSecret)
				c.Difficulty++

				return c
			},
			args:    validSecret,
			wantErr: ErrInvalidSignature,
		},
		{
			name: "updated unix timestamp",
			makeChallenge: func() *Challenge {
				c := NewChallenge(rand, dif, unix, validSecret)
				c.UnixTimestamp++

				return c
			},
			args:    validSecret,
			wantErr: ErrInvalidSignature,
		},
		{
			name: "updated  signature",
			makeChallenge: func() *Challenge {
				c := NewChallenge(rand, dif, unix, validSecret)
				c.Signature = append(c.Signature, 0x01)

				return c
			},
			args:    validSecret,
			wantErr: ErrInvalidSignature,
		},
		{
			name: "set empty signature",
			makeChallenge: func() *Challenge {
				c := NewChallenge(rand, dif, unix, validSecret)
				c.Signature = nil

				return c
			},
			args:    validSecret,
			wantErr: ErrInvalidSignature,
		},
		{
			name: "updated  nonce",
			makeChallenge: func() *Challenge {
				c := NewChallenge(rand, 1, time.Now().Unix(), validSecret)
				c.Nonce++

				return c
			},
			args:    validSecret,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := tt.makeChallenge()

			err := c.VerifySign(tt.args)

			assert.ErrorIsf(t, err, tt.wantErr, "VerifySign() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestChallenge_ComputeNonce(t *testing.T) {
	t.Parallel()

	var (
		secret = []byte("secret")
		rand   = []byte{0x01, 0x02, 0x03, 0x04}
		unix   = time.Now().Unix()
	)

	tests := []struct {
		wantErr       error
		makeChallenge func() *Challenge
		makeCtx       func() context.Context
		name          string
		wantNonce     int64
	}{
		{
			name: "difficulty 0",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 0, unix, secret)
			},
			makeCtx: func() context.Context {
				return context.Background()
			},
			wantErr:   nil,
			wantNonce: 0,
		},
		{
			name: "difficulty 1",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 1, unix, secret)
			},
			makeCtx: func() context.Context {
				return context.Background()
			},
			wantErr:   nil,
			wantNonce: 1,
		},
		{
			name: "difficulty 2",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 2, unix, secret)
			},
			makeCtx: func() context.Context {
				return context.Background()
			},
			wantErr:   nil,
			wantNonce: 2,
		},
		{
			name: "difficulty 3",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 3, unix, secret)
			},
			makeCtx: func() context.Context {
				return context.Background()
			},
			wantErr:   nil,
			wantNonce: 2,
		},
		{
			name: "difficulty 4",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 4, unix, secret)
			},
			makeCtx: func() context.Context {
				return context.Background()
			},
			wantErr:   nil,
			wantNonce: 18,
		},
		{
			name: "difficulty 10",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 10, unix, secret)
			},
			makeCtx: func() context.Context {
				return context.Background()
			},
			wantErr:   nil,
			wantNonce: 252,
		},
		{
			name: "difficulty 20",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 20, unix, secret)
			},
			makeCtx: func() context.Context {
				return context.Background()
			},
			wantErr:   nil,
			wantNonce: 1829638,
		},
		{
			name: "difficulty 255",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 255, unix, secret)
			},
			makeCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			},
			wantErr: context.Canceled,
		},
		{
			name: "context cancelled",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 10, unix, secret)
			},
			makeCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			},
			wantErr: context.Canceled,
		},
		{
			name: "context deadline exceeded",
			makeChallenge: func() *Challenge {
				return NewChallenge(rand, 10, unix, secret)
			},
			makeCtx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 0)
				cancel()

				return ctx
			},
			wantErr: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := tt.makeChallenge()
			err := c.ComputeNonce(tt.makeCtx())

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantNonce, c.Nonce)
		})
	}
}

func TestChallenge_VerifyNonce(t *testing.T) {
	t.Parallel()

	var (
		rand       = []byte{0x01, 0x02, 0x03, 0x04}
		dif  uint8 = 3
	)

	tests := []struct {
		wantErr       error
		makeChallenge func() *Challenge
		name          string
	}{
		{
			name: "valid nonce",
			makeChallenge: func() *Challenge {
				return &Challenge{
					Rand:       rand,
					Nonce:      2,
					Difficulty: dif,
				}
			},
			wantErr: nil,
		},
		{
			name: "invalid nonce",
			makeChallenge: func() *Challenge {
				return &Challenge{
					Rand:       rand,
					Nonce:      15,
					Difficulty: dif,
				}
			},
			wantErr: ErrInvalidNonce,
		},
		{
			name: "update rand",
			makeChallenge: func() *Challenge {
				return &Challenge{
					Rand:       append(rand, 0x01),
					Nonce:      2,
					Difficulty: dif,
				}
			},
			wantErr: ErrInvalidNonce,
		},
		{
			name: "decrease difficulty",
			makeChallenge: func() *Challenge {
				dif := dif - 1

				return &Challenge{
					Rand:       rand,
					Nonce:      2,
					Difficulty: dif,
				}
			},
			wantErr: nil,
		},
		{
			name: "increase difficulty",
			makeChallenge: func() *Challenge {
				dif := dif + 1

				return &Challenge{
					Rand:       rand,
					Nonce:      2,
					Difficulty: dif,
				}
			},
			wantErr: ErrInvalidNonce,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := tt.makeChallenge()

			err := c.VerifyNonce()

			assert.ErrorIsf(t, err, tt.wantErr, "VerifySign() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
