package lstesting

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/amterp/rad/radls/lsp"
	"github.com/amterp/rad/radls/rpc"
	"github.com/amterp/rad/radls/server"
)

const testURI = "file:///test.rad"

// Run executes a single snapshot test case against a real in-process server.
// It handles the full LSP lifecycle: initialization, didOpen, actions, and shutdown.
// Returns the normalized JSON output of all server messages after the init handshake.
func Run(tc *SnapshotCase) (string, error) {
	serverReader, clientWriter := io.Pipe()
	clientReader, serverWriter := io.Pipe()

	s := server.NewServer(serverReader, serverWriter)

	// Server goroutine: runs the full Mux lifecycle, closes its writer on exit
	// so the collector sees EOF. Recovers panics so they surface as test errors
	// rather than misleading timeouts.
	serverDone := make(chan error, 1)
	go func() {
		defer serverWriter.Close()
		defer func() {
			if r := recover(); r != nil {
				serverDone <- fmt.Errorf("server panic: %v\n%s", r, debug.Stack())
			}
		}()
		serverDone <- s.Run()
	}()

	// Collector goroutine: reads raw JSON-RPC messages from the server's output
	// concurrently. This is required because io.Pipe has zero internal buffering -
	// if we don't read while the server writes, the server blocks.
	var collectedMessages []json.RawMessage
	var collectErr error
	collectDone := make(chan struct{})
	go func() {
		defer close(collectDone)
		collectedMessages, collectErr = collectRawMessages(bufio.NewReader(clientReader))
	}()

	// Main goroutine: send messages to the server
	bw := bufio.NewWriter(clientWriter)

	if err := sendInitialize(bw); err != nil {
		clientWriter.Close()
		return "", fmt.Errorf("failed to send initialize: %w", err)
	}

	if err := sendDidOpen(bw, tc.Document); err != nil {
		clientWriter.Close()
		return "", fmt.Errorf("failed to send didOpen: %w", err)
	}

	version := 2
	requestId := 1
	for _, action := range tc.Actions {
		var err error
		switch action.Type {
		case ActionChange:
			err = sendDidChange(bw, action.Content, version)
			version++
		case ActionCompletion:
			err = sendCompletion(bw, requestId, *action.Position)
			requestId++
		case ActionCodeAction:
			err = sendCodeAction(bw, requestId, *action.Range)
			requestId++
		}
		if err != nil {
			clientWriter.Close()
			return "", fmt.Errorf("failed to send action: %w", err)
		}
	}

	// Close the writer to signal EOF. The server's rpc.Decode will return an error,
	// causing Mux.Run() to exit, which closes serverWriter, which causes the
	// collector to see EOF.
	clientWriter.Close()

	select {
	case <-serverDone:
	case <-time.After(5 * time.Second):
		// Close pipes to unblock any goroutines stuck on pipe I/O
		clientReader.Close()
		return "", fmt.Errorf("server did not shut down within timeout")
	}

	select {
	case <-collectDone:
	case <-time.After(2 * time.Second):
		clientReader.Close()
		return "", fmt.Errorf("collector did not finish within timeout")
	}

	if collectErr != nil {
		return "", fmt.Errorf("failed to collect output: %w", collectErr)
	}

	// The first message is the initialize response - skip it.
	if len(collectedMessages) > 0 {
		collectedMessages = collectedMessages[1:]
	}

	return normalizeMessages(collectedMessages)
}

// collectRawMessages reads JSON-RPC messages from a buffered reader until EOF.
// It parses Content-Length headers to extract raw JSON bodies.
// We can't use rpc.Decode here because lsp.IncomingMsg lacks the 'result' field
// that server responses contain.
func collectRawMessages(br *bufio.Reader) ([]json.RawMessage, error) {
	var messages []json.RawMessage
	tr := textproto.NewReader(br)

	for {
		header, err := tr.ReadMIMEHeader()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Non-EOF errors on the header read indicate the pipe was closed
			// mid-stream or the server exited. Since we're reading from an
			// io.Pipe, the server closing its writer appears as an EOF-like
			// error from textproto. Treat any error here as end of messages.
			break
		}

		contentLenStr := header.Get("Content-Length")
		contentLen, err := strconv.ParseInt(contentLenStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length '%s': %w", contentLenStr, err)
		}

		body := make([]byte, contentLen)
		if _, err := io.ReadFull(br, body); err != nil {
			return nil, fmt.Errorf("failed to read message body: %w", err)
		}

		messages = append(messages, json.RawMessage(body))
	}

	return messages, nil
}

// sendRequest sends a JSON-RPC request with the given id, method, and params.
func sendRequest(bw *bufio.Writer, id int, method string, params any) error {
	rawId := marshalRaw(id)
	rawParams := marshalRaw(params)

	msg := lsp.Request{
		Msg:    lsp.Msg{Rpc: "2.0"},
		Id:     &rawId,
		Method: method,
		Params: &rawParams,
	}
	return rpc.Encode(bw, msg)
}

// sendNotification sends a JSON-RPC notification with the given method and params.
func sendNotification(bw *bufio.Writer, method string, params any) error {
	rawParams := marshalRaw(params)

	msg := lsp.Notification{
		Msg:    lsp.Msg{Rpc: "2.0"},
		Method: method,
		Params: &rawParams,
	}
	return rpc.Encode(bw, msg)
}

func sendInitialize(bw *bufio.Writer) error {
	params := lsp.InitializeParams{
		ClientInfo: &lsp.ClientInfo{
			Name:    "test-client",
			Version: "0.0.0",
		},
	}
	return sendRequest(bw, 0, lsp.INITIALIZE, params)
}

func sendDidOpen(bw *bufio.Writer, document string) error {
	params := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			Uri:        testURI,
			LanguageId: "rad",
			Version:    1,
			Text:       document,
		},
	}
	return sendNotification(bw, lsp.TD_DID_OPEN, params)
}

func sendDidChange(bw *bufio.Writer, content string, version int) error {
	params := lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{Uri: testURI},
			Version:                version,
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{
			{Text: content},
		},
	}
	return sendNotification(bw, lsp.TD_DID_CHANGE, params)
}

func sendCompletion(bw *bufio.Writer, id int, pos lsp.Pos) error {
	params := lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{Uri: testURI},
			Position:     pos,
		},
	}
	return sendRequest(bw, id, lsp.TD_COMPLETION, params)
}

func sendCodeAction(bw *bufio.Writer, id int, r lsp.Range) error {
	params := lsp.CodeActionParams{
		TextDocument: lsp.TextDocumentIdentifier{Uri: testURI},
		Range:        r,
	}
	return sendRequest(bw, id, lsp.TD_CODE_ACTION, params)
}

// marshalRaw marshals a value into a json.RawMessage.
func marshalRaw(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal: %v", err))
	}
	return json.RawMessage(b)
}
