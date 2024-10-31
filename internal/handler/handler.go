package handler

import (
	"context"
	"math/rand"
	"time"
)

type QuoteHandler struct {
	quotes []string
}

func NewQuoteHandler(quotes []string) *QuoteHandler {
	return &QuoteHandler{quotes: quotes}
}

// RandomQuote is a demonstration handler
// not used context here is for demo purposes only, to show that it is best practice for building request handlers.
// nolint: gosec //no need crypto safety
func (q *QuoteHandler) RandomQuote(_ context.Context) ([]byte, error) {
	return []byte(q.quotes[rand.Intn(len(q.quotes))]), nil
}

// CurrentTime is a demonstration handler
// not used context here is for demo purposes only, to show that it is best practice for building request handlers.
func (q *QuoteHandler) CurrentTime(_ context.Context) ([]byte, error) {
	return []byte(time.Now().Format(time.DateTime)), nil
}
