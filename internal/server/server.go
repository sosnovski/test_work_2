package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/sosnovski/test_work_2/internal/pow"
	"github.com/sosnovski/test_work_2/internal/proto"
	"go.uber.org/zap"
)

var (
	ErrOddArguments         = errors.New("odd number of arguments")
	errInternalServerError  = errors.New("internal server error")
	errChallengeWasExpired  = errors.New("challenge was expired")
	errChallengeNotFound    = errors.New("challenge not found")
	errUndefinedRequestType = errors.New("undefined request type")
	errHandlerNotFound      = errors.New("handler not found")
)

type (
	StopFunc func(context.Context) error
	Handler  func(context.Context) ([]byte, error)
)

type Server struct {
	log                     *zap.Logger
	handlers                sync.Map
	cache                   *bigcache.BigCache
	secret                  []byte
	wg                      sync.WaitGroup
	readTimeout             time.Duration
	writeTimeout            time.Duration
	challengeTimeout        time.Duration
	challengeRandBytesCount int
	powDifficulty           uint8
}

func NewServer(
	log *zap.Logger,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	powDifficulty uint8,
	challengeTimeout time.Duration,
	challengeRandBytesCount int,
	secret []byte,
	cache *bigcache.BigCache,
) *Server {
	return &Server{
		log:                     log,
		readTimeout:             readTimeout,
		writeTimeout:            writeTimeout,
		powDifficulty:           powDifficulty,
		challengeTimeout:        challengeTimeout,
		challengeRandBytesCount: challengeRandBytesCount,
		secret:                  secret,
		cache:                   cache,
	}
}

func (s *Server) RegisterHandler(resourceID proto.ResourceIDType, handler Handler) {
	s.handlers.Store(resourceID, handler)
}

func (s *Server) RegisterHandlers(resourcesAndFuncs ...interface{}) error {
	if len(resourcesAndFuncs)%2 != 0 {
		return ErrOddArguments //nolint: wrapcheck //nothing to wrap
	}

	for index := 0; index <= len(resourcesAndFuncs)-2; {
		funcIndex := index + 1
		resource := resourcesAndFuncs[index]
		fun := resourcesAndFuncs[funcIndex]

		resourceID, ok := resource.(proto.ResourceIDType)
		if !ok {
			return fmt.Errorf("%d argument must be ResourceIDType, got, %T", index, resource)
		}

		handler, ok := fun.(func(context.Context) ([]byte, error))
		if !ok {
			return fmt.Errorf("%d argument must be Handler, got %T", funcIndex, fun)
		}

		s.RegisterHandler(resourceID, handler)

		index += 2
	}

	return nil
}

// Listen non-blocking method to start listed TCP.
func (s *Server) Listen(ctx context.Context, addr string) (StopFunc, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", addr, err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := listener.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					s.log.Error("accept connection", zap.Error(err))
				}

				continue
			}

			s.wg.Add(1)
			go s.handleConnection(ctx, conn)
		}
	}()

	s.log.Info("server started", zap.String("addr", addr))

	return s.stopFunc(listener), nil
}

func (s *Server) stopFunc(listener net.Listener) StopFunc {
	return func(ctx context.Context) error {
		stopCh := make(chan error, 1)

		defer s.log.Info("start stopped")

		go func() {
			defer close(stopCh)

			err := listener.Close()

			s.wg.Wait()

			if err != nil {
				stopCh <- err
			}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-stopCh:
			return err
		}
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	s.log.Info("start handle new connection", zap.String("remote", conn.RemoteAddr().String()))

	defer func() {
		if err := conn.Close(); err != nil {
			s.log.Error("close connection", zap.Error(err))
		}

		s.wg.Done()

		s.log.Info("stop handle connection", zap.String("remote", conn.RemoteAddr().String()))
	}()

	s.handleRequests(ctx, conn, s.startReadLoop(conn))
}

func (s *Server) startReadLoop(conn net.Conn) <-chan proto.Request {
	messagesCh := make(chan proto.Request)

	go func() {
		defer close(messagesCh)

		for {
			s.setReadDeadline(conn)

			message, err := proto.ReadRequest(conn)
			if err != nil {
				if errors.Is(err, proto.ErrConnectionClosed) ||
					errors.Is(err, proto.ErrDeadlineExceeded) {
					return
				}

				s.log.Error("read request", zap.Error(fmt.Errorf("read message: %w", err)))

				continue
			}

			messagesCh <- *message
		}
	}()

	return messagesCh
}

func (s *Server) writeResponse(conn net.Conn, res proto.Response) {
	s.setWriteDeadline(conn)

	if err := proto.WriteResponse(conn, res); err != nil {
		s.log.Error("write challenge response", zap.Error(err))
	}
}

func (s *Server) handleChallengeRequest() (*proto.Response, error) {
	randBytes := make([]byte, s.challengeRandBytesCount)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, fmt.Errorf("read random bytes: %w", err)
	}

	challenge := pow.NewChallenge(randBytes, s.powDifficulty, time.Now().Unix(), s.secret)

	if s.cache != nil {
		if err := s.cache.Set(string(randBytes), nil); err != nil {
			return nil, fmt.Errorf("cache set: %w", err)
		}
	}

	b, err := json.Marshal(challenge)
	if err != nil {
		return nil, fmt.Errorf("marshal challenge response: %w", err)
	}

	return proto.OkResponse(b), nil
}

func (s *Server) handleResourceRequest(ctx context.Context, request proto.Request) (*proto.Response, error) {
	challenge := &pow.Challenge{}
	if err := json.Unmarshal(request.Payload, challenge); err != nil {
		return nil, fmt.Errorf("unmarshal challenge response: %w", err)
	}

	if s.cache != nil {
		_, err := s.cache.Get(string(challenge.Rand))
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return proto.ErrorResponse(errChallengeNotFound), nil
		}

		if err != nil {
			s.log.Error("get challenge from cache", zap.Error(err))
		}

		if err := s.cache.Delete(string(challenge.Rand)); err != nil {
			s.log.Error("delete challenge from cache", zap.Error(err))
		}
	}

	if err := challenge.VerifySign(s.secret); err != nil {
		return proto.ErrorResponse(fmt.Errorf("invalid sign: %w", err)), nil
	}

	if challenge.UnixTimestamp < time.Now().Add(-s.challengeTimeout).Unix() {
		return proto.ErrorResponse(errChallengeWasExpired), nil
	}

	if err := challenge.VerifyNonce(); err != nil {
		return proto.ErrorResponse(fmt.Errorf("invalid nonce: %w", err)), nil
	}

	value, ok := s.handlers.Load(request.ResourceID)
	if !ok {
		return proto.ErrorResponse(errHandlerNotFound), nil
	}

	handler, ok := value.(Handler)
	if !ok {
		return nil, fmt.Errorf("handler not implemented Handler type: %T", value)
	}

	data, err := handler(ctx)
	if err != nil {
		return proto.ErrorResponse(err), nil
	}

	return proto.OkResponse(data), nil
}

func (s *Server) setReadDeadline(conn net.Conn) {
	if s.readTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(s.readTimeout)); err != nil {
			s.log.Error("set read deadline", zap.Error(err))
		}
	}
}

func (s *Server) setWriteDeadline(conn net.Conn) {
	if s.writeTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(s.writeTimeout)); err != nil {
			s.log.Error("set write deadline", zap.Error(err))
		}
	}
}

func (s *Server) handleRequests(ctx context.Context, conn net.Conn, ch <-chan proto.Request) {
	for {
		select {
		case <-ctx.Done():
			return
		case request := <-ch:
			var (
				response    *proto.Response
				err         error
				ctx, cancel = context.WithCancel(ctx)
			)

			switch request.Type {
			case proto.RequestTypeExit:
				cancel()

				return
			case proto.RequestTypeChallenge:
				response, err = s.handleChallengeRequest()
			case proto.RequestTypeResource:
				response, err = s.handleResourceRequest(ctx, request)
			default:
				response = proto.ErrorResponse(errUndefinedRequestType)
			}

			if err != nil {
				s.log.Error("handle request", zap.Error(err), zap.Int8("type", int8(request.Type)))

				response = proto.ErrorResponse(errInternalServerError)
			}

			s.writeResponse(conn, *response)

			cancel()
		}
	}
}
