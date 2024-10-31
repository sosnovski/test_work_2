package proto

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
)

var (
	ErrConnectionClosed     = errors.New("connection is closed")
	ErrDeadlineExceeded     = errors.New("deadline exceeded")
	ErrInvalidPayloadLength = errors.New("invalid payload length")
	ErrZeroBytesWritten     = errors.New("zero bytes written")
	ErrReadAction           = errors.New("read action error")
	ErrReadStatus           = errors.New("read status error")
	ErrReadPayload          = errors.New("read payload error")
	ErrReadResourceID       = errors.New("read resource ID error")
	ErrInvalidBytesCount    = errors.New("invalid bytes count")
)

func wrapWriteErr(err error) error {
	if errors.Is(err, syscall.EPIPE) {
		return fmt.Errorf("%w: %w", ErrConnectionClosed, err)
	}

	return err
}

func wrapReadErr(err error) error {
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return fmt.Errorf("%w: %w", ErrDeadlineExceeded, err)
	}

	if errors.Is(err, net.ErrClosed) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, syscall.ECONNRESET) {
		return fmt.Errorf("%w: %w", ErrConnectionClosed, err)
	}

	return err
}
