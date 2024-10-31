package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type (
	RequestType    byte
	StatusType     byte
	ResourceIDType uint16
)

const (
	RequestTypeExit RequestType = iota
	RequestTypeChallenge
	RequestTypeResource
)

const (
	StatusOK StatusType = iota
	StatusErr
)

const (
	actionBytesCount        = 1 // bytes count allocated for transmitting an action type (response or response)
	statusBytesCount        = 1 // bytes count allocated for transmitting a status
	resourceIDBytesCount    = 2 // bytes count allocated for transmitting a resource ID
	contentLengthBytesCount = 4 // bytes count allocated for transmitting a content length

	maxPayloadLength = math.MaxUint32
)

type Request struct {
	Payload    []byte
	ResourceID ResourceIDType
	Type       RequestType
}

type Response struct {
	Payload []byte
	Status  StatusType
}

func MakeResponse(status StatusType, payload []byte) *Response {
	return &Response{
		Status:  status,
		Payload: payload,
	}
}

func ErrorResponse(err error) *Response {
	return MakeResponse(StatusErr, []byte(err.Error()))
}

func OkResponse(data []byte) *Response {
	return MakeResponse(StatusOK, data)
}

func ReadRequest(reader io.Reader) (*Request, error) {
	action, err := readBytes(reader, actionBytesCount)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadAction, err)
	}

	resourceID, err := readResourceID(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadResourceID, err)
	}

	payload, err := readPayload(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadPayload, err)
	}

	return &Request{
		Type:       RequestType(action[0]),
		ResourceID: ResourceIDType(resourceID),
		Payload:    payload,
	}, nil
}

func ReadResponse(reader io.Reader) (*Response, error) {
	status, err := readBytes(reader, statusBytesCount)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadStatus, err)
	}

	payload, err := readPayload(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadPayload, err)
	}

	return &Response{
		Status:  StatusType(status[0]),
		Payload: payload,
	}, nil
}

func WriteRequest(writer io.Writer, req Request) error {
	length := len(req.Payload)
	if length > maxPayloadLength {
		return fmt.Errorf("%w, max length %d: %d", ErrInvalidPayloadLength, maxPayloadLength, length)
	}

	resourceIDBytes := make([]byte, resourceIDBytesCount)
	binary.LittleEndian.PutUint16(resourceIDBytes, uint16(req.ResourceID))

	lengthBytes := make([]byte, contentLengthBytesCount)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(length)) //nolint:gosec //overflow was checked on prev lines

	buffer := bytes.NewBuffer([]byte{byte(req.Type)})
	buffer.Write(resourceIDBytes)
	buffer.Write(lengthBytes)
	buffer.Write(req.Payload)

	if err := writeAndWrapErr(writer, buffer.Bytes()); err != nil {
		return err
	}

	return nil
}

func WriteResponse(writer io.Writer, res Response) error {
	length := len(res.Payload)
	if length > maxPayloadLength {
		return fmt.Errorf("%w, max length %d: %d", ErrInvalidPayloadLength, maxPayloadLength, length)
	}

	lengthBytes := make([]byte, contentLengthBytesCount)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(length)) //nolint:gosec //overflow was checked on prev lines

	buffer := bytes.NewBuffer([]byte{byte(res.Status)})
	buffer.Write(lengthBytes)
	buffer.Write(res.Payload)

	if err := writeAndWrapErr(writer, buffer.Bytes()); err != nil {
		return err
	}

	return nil
}

func readPayload(reader io.Reader) ([]byte, error) {
	contentLengthBytes, err := readBytes(reader, contentLengthBytesCount)
	if err != nil {
		return nil, fmt.Errorf("read content length bytes: %w", err)
	}

	contentLength := binary.LittleEndian.Uint32(contentLengthBytes)

	if contentLength > 0 {
		payload, err := readBytes(reader, int(contentLength))
		if err != nil {
			return nil, fmt.Errorf("read payload bytes: %w", err)
		}

		return payload, nil
	}

	return nil, nil
}

func readResourceID(reader io.Reader) (uint16, error) {
	resourceIDBytes, err := readBytes(reader, resourceIDBytesCount)
	if err != nil {
		return 0, fmt.Errorf("read resource id bytes: %w", err)
	}

	return binary.LittleEndian.Uint16(resourceIDBytes), nil
}

func readBytes(reader io.Reader, needRead int) ([]byte, error) {
	data := make([]byte, needRead)

	readCount, err := reader.Read(data)
	if err != nil {
		return nil, wrapReadErr(err)
	}

	if readCount != needRead {
		return nil, fmt.Errorf("%w: %d", ErrInvalidBytesCount, readCount)
	}

	return data, nil
}

func writeAndWrapErr(writer io.Writer, p []byte) error {
	n, err := writer.Write(p)
	if err != nil {
		return wrapWriteErr(err)
	}

	if n == 0 {
		return ErrZeroBytesWritten
	}

	return nil
}
