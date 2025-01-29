package lsp

import (
	"encoding/json"
)

func NewResponseError(id *json.RawMessage, err error) (resp Response) {
	return Response{
		Msg: Msg{
			Rpc: "2.0",
		},
		Id:     id,
		Result: nil,
		Error:  newError(err),
	}
}

func newError(err error) *Error {
	if err != nil {
		return nil
	}
	return &Error{
		Code: 0, // todo should probably not be 0
		Msg:  err.Error(),
		Data: nil,
	}
}
