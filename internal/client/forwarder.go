package client

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/prabhatlabs/go-tunnel/internal/protocol"
)

func httpReqForwarder(port int, req *protocol.RequestMessage) (*protocol.ResponseMessage, error) {
	baseUrl := "http://localhost:" + strconv.Itoa(port)
	url := baseUrl + req.Path
	client := &http.Client{}

	resp := &protocol.ResponseMessage{
		ID:      req.ID,
		Type:    protocol.MsgTypeResponse,
		Status:  http.StatusBadGateway,
		Headers: make(map[string]string),
		Body:    nil,
	}

	httpReq, err := http.NewRequest(req.Method, url, bytes.NewReader(req.Body))
	if err != nil {
		return resp, err
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return resp, err
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return resp, err
	}

	resp.Status = httpResp.StatusCode
	resp.Body = body
	for key, values := range httpResp.Header {
		resp.Headers[key] = values[0]
	}

	return resp, nil
}
