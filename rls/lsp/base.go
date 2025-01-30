package lsp

import "encoding/json"

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#message
type Msg struct {
	Rpc string `json:"jsonrpc"` // probably always be 2.0
}
type IncomingMsg struct {
	Msg
	Id     *json.RawMessage `json:"id"`
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
}

func (m IncomingMsg) IsRequest() bool {
	return m.Id != nil
}

func (m IncomingMsg) IsNotification() bool {
	return m.Id == nil
}

func (m IncomingMsg) AsRequest() (Request, bool) {
	if !m.IsRequest() {
		return Request{}, false
	}
	return Request{
		Msg:    m.Msg,
		Id:     m.Id,
		Method: m.Method,
		Params: m.Params,
	}, true
}

func (m IncomingMsg) AsNotification() (Notification, bool) {
	if !m.IsNotification() {
		return Notification{}, false
	}
	return Notification{
		Msg:    m.Msg,
		Method: m.Method,
		Params: m.Params,
	}, true
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#notificationMessage
type Notification struct {
	Msg
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#requestMessage
type Request struct {
	Msg
	Id     *json.RawMessage `json:"id"`
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#responseMessage
type Response struct {
	Msg
	Id     *json.RawMessage `json:"id"`
	Result any              `json:"result"`
	Error  *Error           `json:"error,omitempty"`
}

func NewResponse(id *json.RawMessage, result any) Response {
	return Response{
		Msg: Msg{
			Rpc: "2.0",
		},
		Id:     id,
		Result: result,
		Error:  nil,
	}
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#responseError
type Error struct {
	Code int64 `json:"code"`
	// Should be limited to a single concise sentence.
	Msg  string `json:"message"`
	Data any    `json:"data,omitempty"`
}
