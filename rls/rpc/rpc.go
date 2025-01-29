package rpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"rls/com"
	"rls/lsp"
	"strconv"
)

func Decode(r *bufio.Reader) (msg lsp.IncomingMsg, err error) {
	header, err := textproto.NewReader(r).ReadMIMEHeader()
	if err != nil {
		return msg, com.ErrFailedReadingHeader
	}

	contentLen, err := strconv.ParseInt(header.Get("Content-Length"), 10, 64)
	if err != nil {
		return msg, com.ErrInvalidContentLengthHeader
	}

	err = json.NewDecoder(io.LimitReader(r, contentLen)).Decode(&msg)
	if err != nil {
		return msg, com.ErrFailedDecode
	}
	return
}

func Encode(w *bufio.Writer, msg any) (err error) {
	content, err := json.Marshal(msg)
	if err != nil {
		return
	}

	headers := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))

	if _, err = w.WriteString(headers); err != nil {
		return
	}
	if _, err = w.Write(content); err != nil {
		return
	}
	return w.Flush()
}
