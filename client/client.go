package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/sosnovski/test_work_2/internal/pow"
	"github.com/sosnovski/test_work_2/internal/proto"
)

var (
	ErrResponse             = errors.New("response error")
	ErrUnsupportedValueType = errors.New("unsupported value type")
	ErrNotPointerValue      = errors.New("not a pointer value")
)

const (
	QuoteResourceID proto.ResourceIDType = 0
	TimeResourceID  proto.ResourceIDType = 1
)

type Client struct {
	conn                    net.Conn
	address                 string
	challengeComputeTimeout time.Duration
	requestTimeout          time.Duration
	mu                      sync.Mutex
}

func New(
	address string,
	challengeComputeTimeout time.Duration,
	requestTimeout time.Duration,
) *Client {
	return &Client{
		address:                 address,
		challengeComputeTimeout: challengeComputeTimeout,
		requestTimeout:          requestTimeout,
	}
}

func (c *Client) Quote(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	var quote string
	if err := c.doChallenge(ctx, QuoteResourceID, &quote); err != nil {
		return "", fmt.Errorf("failed to do challenge: %w", err)
	}

	return quote, nil
}

func (c *Client) CurrentTime(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	var currentTime string
	if err := c.doChallenge(ctx, TimeResourceID, &currentTime); err != nil {
		return "", fmt.Errorf("failed to do challenge: %w", err)
	}

	return currentTime, nil
}

func (c *Client) establishConn(ctx context.Context) error {
	if c.conn == nil {
		conn, err := new(net.Dialer).DialContext(ctx, "tcp", c.address)
		if err != nil {
			return fmt.Errorf("connect to server: %w", err)
		}

		c.conn = conn
	}

	return nil
}

func (c *Client) doChallenge(ctx context.Context, resourceID proto.ResourceIDType, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.establishConn(ctx)
	if err != nil {
		return fmt.Errorf("establish connection: %w", err)
	}

	challenge := &pow.Challenge{}

	if err := c.request(ctx, request{Type: proto.RequestTypeChallenge}, challenge); err != nil {
		return fmt.Errorf("request challenge: %w", err)
	}

	computeCtx := ctx

	if c.challengeComputeTimeout > 0 {
		var cancel context.CancelFunc

		computeCtx, cancel = context.WithTimeout(ctx, c.challengeComputeTimeout)
		defer cancel()
	}

	if err := challenge.ComputeNonce(computeCtx); err != nil {
		return fmt.Errorf("compute challenge: %w", err)
	}

	if err := c.request(ctx, request{
		Type:       proto.RequestTypeResource,
		ResourceID: resourceID,
		Data:       challenge,
	}, value); err != nil {
		return fmt.Errorf("request resource: %w", err)
	}

	return nil
}

func (c *Client) request(ctx context.Context, request request, value any) error {
	err := c.do(ctx, request, value)

	if errors.Is(err, proto.ErrConnectionClosed) {
		if closeErr := c.conn.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close connection: %w", closeErr))
		}

		c.conn = nil
	}

	return err //nolint: wrapcheck //nothing to wrap
}

func (c *Client) do(ctx context.Context, request request, value any) error {
	deadline, ok := ctx.Deadline()
	if ok {
		if err := c.conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("set request deadline: %w", err)
		}
	}

	requestMessage, err := request.makeRequest()
	if err != nil {
		return fmt.Errorf("make message: %w", err)
	}

	if err := proto.WriteRequest(c.conn, *requestMessage); err != nil {
		return fmt.Errorf("write request %d: %w", requestMessage.Type, err)
	}

	responseMessage, err := proto.ReadResponse(c.conn)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if responseMessage.Status != proto.StatusOK {
		return fmt.Errorf("%w: %s, %d", ErrResponse, responseMessage.Payload, responseMessage.Status)
	}

	if value != nil {
		reflectVal := reflect.ValueOf(value)

		if reflectVal.Kind() != reflect.Pointer {
			return ErrNotPointerValue //nolint: wrapcheck //nothing to wrap
		}

		kind := reflectVal.Elem().Kind()

		switch kind { //nolint: exhaustive //used default state
		case reflect.String:
			stringValue, ok := value.(*string)
			if !ok {
				return fmt.Errorf("value must be a string, got %T", value)
			}

			*stringValue = string(responseMessage.Payload)
		case reflect.Struct:
			if err := json.Unmarshal(responseMessage.Payload, value); err != nil {
				return fmt.Errorf("unmarshal response: %w", err)
			}
		default:
			return fmt.Errorf("%w: %s", ErrUnsupportedValueType, reflectVal.Type().String())
		}
	}

	return nil
}
