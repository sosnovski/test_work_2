package client

import (
	"encoding/json"
	"fmt"

	"github.com/sosnovski/test_work_2/internal/proto"
)

type request struct {
	Data       any
	ResourceID proto.ResourceIDType
	Type       proto.RequestType
}

func (r *request) makeRequest() (*proto.Request, error) {
	req := &proto.Request{
		Type: r.Type,
	}

	if r.Data != nil {
		data, err := json.Marshal(r.Data)
		if err != nil {
			return nil, fmt.Errorf("marshaling data: %w", err)
		}

		req.Payload = data
	}

	return req, nil
}
