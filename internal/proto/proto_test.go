package proto

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=proto_test.go -destination=proto_mocks_test.go -package=proto
type writer interface { //nolint: unused //it uses for mock
	io.Writer
}

type reader interface { //nolint: unused //it uses for mock
	io.Reader
}

func TestWriteRequest(t *testing.T) {
	t.Parallel()

	var (
		payload         = []byte("some payload")
		overflowPayload = make([]byte, maxPayloadLength+1)
		ctrl            = gomock.NewController(t)
	)

	tests := []struct {
		wantErr    error
		makeWriter func() io.Writer
		name       string
		args       Request
	}{
		{
			name: "undefined RequestType",
			args: Request{
				Type:    RequestType(255),
				Payload: nil,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().Write([]byte{0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}).Return(1, nil).Times(1)

				return mock
			},
			wantErr: nil,
		},
		{
			name: "response exit",
			args: Request{
				Type:    RequestTypeExit,
				Payload: nil,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}).Return(1, nil).Times(1)

				return mock
			},
			wantErr: nil,
		},
		{
			name: "response challenge",
			args: Request{
				Type:    RequestTypeChallenge,
				Payload: nil,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().Write([]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}).Return(1, nil).Times(1)

				return mock
			},
			wantErr: nil,
		},
		{
			name: "response resource",
			args: Request{
				Type:    RequestTypeResource,
				Payload: nil,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().Write([]byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}).Return(1, nil).Times(1)

				return mock
			},
			wantErr: nil,
		},
		{
			name: "with payload",
			args: Request{
				Type:    RequestTypeResource,
				Payload: payload,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().
					Write([]byte{0x2, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x73, 0x6f, 0x6d, 0x65, 0x20, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64}).
					Return(1, nil).
					Times(1)

				return mock
			},
			wantErr: nil,
		},
		{
			name: "overflow payload",
			args: Request{
				Type:    RequestTypeResource,
				Payload: overflowPayload,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)

				return mock
			},
			wantErr: ErrInvalidPayloadLength,
		},
		{
			name: "zero bytes write",
			args: Request{
				Type:    RequestTypeResource,
				Payload: payload,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().
					Write([]byte{0x2, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x73, 0x6f, 0x6d, 0x65, 0x20, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64}).
					Return(0, nil).
					Times(1)

				return mock
			},
			wantErr: ErrZeroBytesWritten,
		},
		{
			name: "err broken pipe",
			args: Request{
				Type:    RequestTypeResource,
				Payload: payload,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().
					Write(gomock.Any()).
					Return(0, syscall.EPIPE).
					Times(1)

				return mock
			},
			wantErr: ErrConnectionClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := WriteRequest(tt.makeWriter(), tt.args)

			assert.ErrorIsf(t, err, tt.wantErr, "WriteRequest() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestWriteResponse(t *testing.T) {
	t.Parallel()

	var (
		payload         = []byte("some payload")
		overflowPayload = make([]byte, maxPayloadLength+1)
		ctrl            = gomock.NewController(t)
	)

	tests := []struct {
		wantErr    error
		makeWriter func() io.Writer
		name       string
		args       Response
	}{
		{
			name: "status OK",
			args: Response{
				Status:  StatusOK,
				Payload: nil,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().
					Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0}).
					Return(1, nil).
					Times(1)

				return mock
			},
			wantErr: nil,
		},
		{
			name: "status Err",
			args: Response{
				Status:  StatusErr,
				Payload: nil,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().
					Write([]byte{0x01, 0x0, 0x0, 0x0, 0x0}).
					Return(1, nil).
					Times(1)

				return mock
			},
			wantErr: nil,
		},
		{
			name: "with payload",
			args: Response{
				Status:  StatusOK,
				Payload: payload,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().
					Write([]byte{0x0, 0xc, 0x0, 0x0, 0x0, 0x73, 0x6f, 0x6d, 0x65, 0x20, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64}).
					Return(1, nil).
					Times(1)

				return mock
			},
			wantErr: nil,
		},
		{
			name: "overflow payload",
			args: Response{
				Status:  StatusOK,
				Payload: overflowPayload,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)

				return mock
			},
			wantErr: ErrInvalidPayloadLength,
		},
		{
			name: "zero bytes write",
			args: Response{
				Status:  StatusOK,
				Payload: payload,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().
					Write([]byte{0x0, 0xc, 0x0, 0x0, 0x0, 0x73, 0x6f, 0x6d, 0x65, 0x20, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64}).
					Return(0, nil).
					Times(1)

				return mock
			},
			wantErr: ErrZeroBytesWritten,
		},
		{
			name: "err broken pipe",
			args: Response{
				Status:  StatusOK,
				Payload: payload,
			},
			makeWriter: func() io.Writer {
				mock := NewMockwriter(ctrl)
				mock.EXPECT().
					Write(gomock.Any()).
					Return(0, syscall.EPIPE).
					Times(1)

				return mock
			},
			wantErr: ErrConnectionClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := WriteResponse(tt.makeWriter(), tt.args)

			assert.ErrorIsf(t, err, tt.wantErr, "WriteResponse() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestReadRequest(t *testing.T) {
	t.Parallel()

	var (
		ctrl    = gomock.NewController(t)
		err     = errors.New("some error")
		payload = []byte("some payload")
	)

	tests := []struct {
		wantErr    error
		makeReader func() io.Reader
		request    *Request
		name       string
	}{
		{
			name: "response type exit",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(RequestTypeExit)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, resourceIDBytesCount)).
					Return(2, nil).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).
					Return(4, nil).Times(1)

				return mock
			},
			wantErr: nil,
			request: &Request{
				Type:    RequestTypeExit,
				Payload: nil,
			},
		},
		{
			name: "response type challenge",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(RequestTypeChallenge)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, resourceIDBytesCount)).
					Return(2, nil).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).
					Return(4, nil).Times(1)

				return mock
			},
			wantErr: nil,
			request: &Request{
				Type:    RequestTypeChallenge,
				Payload: nil,
			},
		},
		{
			name: "response type resource",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(RequestTypeResource)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, resourceIDBytesCount)).
					Return(2, nil).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).
					Return(4, nil).Times(1)

				return mock
			},
			wantErr: nil,
			request: &Request{
				Type:    RequestTypeResource,
				Payload: nil,
			},
		},
		{
			name: "err read action",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).Return(0, err).Times(1)

				return mock
			},
			wantErr: ErrReadAction,
			request: nil,
		},
		{
			name: "err read action",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).Return(1, nil).Times(1)

				mock.EXPECT().Read(make([]byte, resourceIDBytesCount)).Return(0, err).Times(1)

				return mock
			},
			wantErr: ErrReadResourceID,
			request: nil,
		},
		{
			name: "err i/o timeout",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).Return(1, os.ErrDeadlineExceeded).Times(1)

				return mock
			},
			wantErr: ErrDeadlineExceeded,
			request: nil,
		},
		{
			name: "err use of closed network connection",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).Return(1, net.ErrClosed).Times(1)

				return mock
			},
			wantErr: ErrConnectionClosed,
			request: nil,
		},
		{
			name: "err io.EOF",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).Return(1, io.EOF).Times(1)

				return mock
			},
			wantErr: ErrConnectionClosed,
			request: nil,
		},
		{
			name: "err connection reset by peer",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).Return(1, syscall.ECONNRESET).Times(1)

				return mock
			},
			wantErr: ErrConnectionClosed,
			request: nil,
		},
		{
			name: "response with payload",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(RequestTypeChallenge)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, resourceIDBytesCount)).
					Return(2, nil).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					copy(b, []byte{0xc, 0x0, 0x0, 0x0})

					return 4, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, len(payload))).DoAndReturn(func(b []byte) (int, error) {
					copy(b, payload)

					return len(payload), nil
				}).Times(1)

				return mock
			},
			wantErr: nil,
			request: &Request{
				Type:    RequestTypeChallenge,
				Payload: payload,
			},
		},
		{
			name: "err read payload bytes",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(RequestTypeChallenge)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, resourceIDBytesCount)).
					Return(2, nil).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					copy(b, []byte{0xc, 0x0, 0x0, 0x0})

					return 4, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, len(payload))).Return(0, err).Times(1)

				return mock
			},
			wantErr: ErrReadPayload,
			request: nil,
		},
		{
			name: "err invalid bytes count",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, actionBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(RequestTypeChallenge)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, resourceIDBytesCount)).
					Return(2, nil).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					copy(b, []byte{0xc, 0x0, 0x0, 0x0})

					return 3, nil
				}).Times(1)

				return mock
			},
			wantErr: ErrInvalidBytesCount,
			request: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			message, err := ReadRequest(tt.makeReader())

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.request, message)
		})
	}
}

func TestReadResponse(t *testing.T) {
	t.Parallel()

	var (
		ctrl    = gomock.NewController(t)
		err     = errors.New("some error")
		payload = []byte("some payload")
	)

	tests := []struct {
		wantErr    error
		makeReader func() io.Reader
		response   *Response
		name       string
	}{
		{
			name: "response OK",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(StatusOK)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).
					Return(4, nil).Times(1)

				return mock
			},
			wantErr: nil,
			response: &Response{
				Status:  StatusOK,
				Payload: nil,
			},
		},
		{
			name: "response ERR",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(StatusErr)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).
					Return(4, nil).Times(1)

				return mock
			},
			wantErr: nil,
			response: &Response{
				Status:  StatusErr,
				Payload: nil,
			},
		},
		{
			name: "err read status",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).Return(0, err).Times(1)

				return mock
			},
			wantErr:  ErrReadStatus,
			response: nil,
		},
		{
			name: "err i/o timeout",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).Return(1, os.ErrDeadlineExceeded).Times(1)

				return mock
			},
			wantErr:  ErrDeadlineExceeded,
			response: nil,
		},
		{
			name: "err use of closed network connection",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).Return(1, net.ErrClosed).Times(1)

				return mock
			},
			wantErr:  ErrConnectionClosed,
			response: nil,
		},
		{
			name: "err io.EOF",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).Return(1, io.EOF).Times(1)

				return mock
			},
			wantErr:  ErrConnectionClosed,
			response: nil,
		},
		{
			name: "err connection reset by peer",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).Return(1, syscall.ECONNRESET).Times(1)

				return mock
			},
			wantErr:  ErrConnectionClosed,
			response: nil,
		},
		{
			name: "response with payload",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(StatusOK)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					copy(b, []byte{0xc, 0x0, 0x0, 0x0})

					return 4, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, len(payload))).DoAndReturn(func(b []byte) (int, error) {
					copy(b, payload)

					return len(payload), nil
				}).Times(1)

				return mock
			},
			wantErr: nil,
			response: &Response{
				Status:  StatusOK,
				Payload: payload,
			},
		},
		{
			name: "err read payload bytes",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(StatusOK)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					copy(b, []byte{0xc, 0x0, 0x0, 0x0})

					return 4, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, len(payload))).Return(0, err).Times(1)

				return mock
			},
			wantErr:  ErrReadPayload,
			response: nil,
		},
		{
			name: "err invalid bytes count",
			makeReader: func() io.Reader {
				mock := NewMockreader(ctrl)
				mock.EXPECT().Read(make([]byte, statusBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					b[0] = byte(StatusOK)

					return 1, nil
				}).Times(1)

				mock.EXPECT().Read(make([]byte, contentLengthBytesCount)).DoAndReturn(func(b []byte) (int, error) {
					copy(b, []byte{0xc, 0x0, 0x0, 0x0})

					return 3, nil
				}).Times(1)

				return mock
			},
			wantErr:  ErrInvalidBytesCount,
			response: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			message, err := ReadResponse(tt.makeReader())

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.response, message)
		})
	}
}
