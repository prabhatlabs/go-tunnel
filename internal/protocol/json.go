package protocol

import (
	"encoding/json"
	"fmt"
)

func (m RequestMessage) MarshalJSON() ([]byte, error) {
	type alias RequestMessage
	m.Type = MsgTypeRequest
	return json.Marshal(alias(m))
}

func (m *RequestMessage) UnmarshalJSON(data []byte) error {
	type alias RequestMessage
	if err := json.Unmarshal(data, (*alias)(m)); err != nil {
		return err
	}
	if m.Type != MsgTypeRequest {
		return fmt.Errorf("invalid type for RequestMessage: %q", m.Type)
	}
	return nil
}

func (m ResponseMessage) MarshalJSON() ([]byte, error) {
	type alias ResponseMessage
	m.Type = MsgTypeResponse
	return json.Marshal(alias(m))
}

func (m *ResponseMessage) UnmarshalJSON(data []byte) error {
	type alias ResponseMessage
	if err := json.Unmarshal(data, (*alias)(m)); err != nil {
		return err
	}
	if m.Type != MsgTypeResponse {
		return fmt.Errorf("invalid type for ResponseMessage: %q", m.Type)
	}
	return nil
}
