package protocol

const (
	MsgTypeRequest  = "request"
	MsgTypeResponse = "response"
)

type RequestMessage struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

type ResponseMessage struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}
