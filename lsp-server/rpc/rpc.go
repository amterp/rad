package rpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"strconv"

	"github.com/amterp/rad/lsp-server/com"
	"github.com/amterp/rad/lsp-server/log"
	"github.com/amterp/rad/lsp-server/lsp"
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

	strMsg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(content), string(content))

	log.L.Infow("Writing message", "msg", strMsg)

	if _, err = w.WriteString(strMsg); err != nil {
		return
	}

	return w.Flush()
}
