package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"

	"github.com/amterp/rad/radls/com"
	"github.com/amterp/rad/radls/log"
	"github.com/amterp/rad/radls/lsp"
	"github.com/amterp/rad/radls/rpc"
)

// NotificationHandler handles a JSON-RPC notification. Notifications
// have no response, so the only return value is an error for logging.
// The context carries cancellation derived from the server's session
// context - if the session shuts down, in-flight notifications get
// cancelled too.
type NotificationHandler func(ctx context.Context, params json.RawMessage) (err error)

// RequestHandler handles a JSON-RPC request. The context carries
// cancellation that the client can flip via $/cancelRequest. Long-
// running handlers should check ctx.Err() at sensible checkpoints
// and bail out with a partial result (or the context's error)
// rather than spending budget on work the client no longer wants.
type RequestHandler func(ctx context.Context, params json.RawMessage) (result any, err error)

type Mux struct {
	reader               *bufio.Reader
	writer               *bufio.Writer
	notificationHandlers map[string]NotificationHandler
	requestHandlers      map[string]RequestHandler
	writeLock            *sync.Mutex

	// baseCtx is the session context; per-request contexts are derived
	// from it. When Run() exits we cancel baseCtx, which propagates
	// down to every in-flight handler.
	baseCtx    context.Context
	baseCancel context.CancelFunc

	// inflight tracks the cancel func for each in-flight REQUEST
	// (notifications aren't trackable; they have no id). Keyed by the
	// raw JSON bytes of the request id so we can compare regardless
	// of whether the client used a number or string id.
	inflightMu sync.Mutex
	inflight   map[string]context.CancelFunc
}

func NewMux(r io.Reader, w io.Writer) *Mux {
	ctx, cancel := context.WithCancel(context.Background())
	mux := Mux{
		reader:               bufio.NewReader(r),
		writer:               bufio.NewWriter(w),
		notificationHandlers: make(map[string]NotificationHandler),
		requestHandlers:      make(map[string]RequestHandler),
		writeLock:            &sync.Mutex{},
		baseCtx:              ctx,
		baseCancel:           cancel,
		inflight:             make(map[string]context.CancelFunc),
	}
	// $/cancelRequest is wired here (not in server.go) because it's a
	// protocol-level concern of the Mux: it has to inspect the
	// inflight registry, which the Mux owns.
	mux.AddNotificationHandler(lsp.CANCEL_REQUEST, mux.handleCancelRequest)
	return &mux
}

func (m *Mux) AddNotificationHandler(method string, handler NotificationHandler) {
	m.notificationHandlers[method] = handler
	log.L.Infof("Registered notification handler for method %s", method)
}

func (m *Mux) AddRequestHandler(method string, handler RequestHandler) {
	m.requestHandlers[method] = handler
	log.L.Infof("Registered request handler for method %s", method)
}

func (m *Mux) Notify(method string, params any) (err error) {
	// Param-less notification: emit no "params" field at all (per
	// JSON-RPC). The earlier "msg = nil" branch was dead - msg got
	// unconditionally overwritten below, so nil-params still went
	// through json.Marshal and produced "params":null on the wire.
	if params == nil {
		return m.write(lsp.NewNotification(method, nil))
	}

	b, err := json.Marshal(params)
	if err != nil {
		return
	}

	raw := json.RawMessage(b)
	return m.write(lsp.NewNotification(method, &raw))
}

func (m *Mux) Init() (err error) {
	log.L.Info("Initializing mux, awaiting initialize msg...")
	for {
		var msg lsp.IncomingMsg
		msg, err = rpc.Decode(m.reader)
		if err != nil {
			return
		}
		if msg.IsNotification() {
			if msg.Method != "exit" {
				log.L.Warnw("Dropping notification sent before initialization", "msg", com.FlatStr(msg))
				continue
			}
			err = m.handleMessage(msg)
			continue
		} else if msg.Method != "initialize" {
			log.L.Warnw("The client sent a request before initialization", "msg", com.FlatStr(msg))
			if err = m.write(lsp.NewResponseError(msg.Id, com.ErrServerNotInitialized)); err != nil {
				return
			}
			continue
		}
		err = m.handleMessage(msg)
		break
	}
	log.L.Info("Initialized mux")
	return
}

func (m *Mux) Run() (err error) {
	defer m.baseCancel() // tear down all in-flight handlers on exit
	for {
		var msg lsp.IncomingMsg
		msg, err = rpc.Decode(m.reader)
		if err != nil {
			return err
		}
		err = m.handleMessage(msg)
		// todo actually do something with error? Send to client?
	}
}

func (m *Mux) handleMessage(msg lsp.IncomingMsg) (err error) {
	log.L.Infof("Received message: %s", com.FlatStr(msg))
	if msg.IsNotification() {
		log.L.Info("Notification")
		notification, _ := msg.AsNotification()
		err = m.handleNotification(notification)
	} else if msg.IsRequest() {
		log.L.Info("Request")
		request, _ := msg.AsRequest()
		err = m.handleRequestResponse(request)
	}
	return
}

func (m *Mux) handleNotification(notification lsp.Notification) (err error) {
	handler, ok := m.notificationHandlers[notification.Method]
	if !ok {
		log.L.Infof("No handler for notification %q", notification.Method)
		return
	}
	// Notifications get the session context directly. They have no
	// id and so can't be individually cancelled - they're either
	// processed or dropped when the session ends.
	var rawParams json.RawMessage
	if notification.Params != nil {
		rawParams = *notification.Params
	}
	err = handler(m.baseCtx, rawParams)
	if err == nil {
		log.L.Infof("Successfully handled notification for %s", notification.Method)
	} else {
		log.L.Errorf("Error handling notification %q: %v", notification.Method, err)
	}
	return
}

func (m *Mux) handleRequestResponse(request lsp.Request) (err error) {
	handler, ok := m.requestHandlers[request.Method]
	if !ok {
		log.L.Errorf("No handler for request %q", request.Method)
		err = m.write(lsp.NewResponseError(request.Id, com.ErrMethodNotFound))
		if err != nil {
			log.L.Errorf("Failed to write writing error response: %v", err)
		}
		return
	}

	// Register the request as in-flight before dispatch. If
	// $/cancelRequest arrives for this id, we'll cancel its context.
	// The defer cancel() pairs with WithCancel - per Go contract,
	// every context.WithCancel must have its cancel func called to
	// release the child node from its parent's tree (otherwise the
	// parent leaks the node until it's itself cancelled, here
	// baseCtx at session shutdown).
	ctx, cancel := context.WithCancel(m.baseCtx)
	defer cancel()
	idKey := requestIDKey(request.Id)
	m.registerInflight(idKey, cancel)
	defer m.unregisterInflight(idKey)

	// LSP requests can legitimately omit `params` (or send null).
	// Don't dereference a nil pointer; hand the handler an empty
	// json.RawMessage and let it Unmarshal as it sees fit.
	var rawParams json.RawMessage
	if request.Params != nil {
		rawParams = *request.Params
	}
	result, err := handler(ctx, rawParams)
	if err != nil {
		log.L.Errorf("Error handling request %q: %v", request.Method, err)
		err = m.write(lsp.NewResponseError(request.Id, err))
		if err != nil {
			log.L.Errorf("Failed to write error response: %v", err)
		}
		return
	}
	log.L.Infow("Sending result", "result", com.FlatStr(result), "method", request.Method)

	resp := lsp.NewResponse(request.Id, result)

	err = m.write(resp)
	if err != nil {
		log.L.Errorf("Failed to write response: %v", err)
		err = m.write(lsp.NewResponseError(request.Id, err))
		if err != nil {
			log.L.Errorf("Failed to write error response: %v", err)
		}
	}
	log.L.Info("Responded.")
	return
}

// handleCancelRequest is the wire-level handler for $/cancelRequest.
// It looks up the inflight registry by request id and fires the
// stored cancel function; the request handler observes ctx.Done()
// at its next checkpoint. If the id is unknown the cancel arrived
// after the request already completed - benign, just log.
func (m *Mux) handleCancelRequest(_ context.Context, params json.RawMessage) error {
	var p lsp.CancelParams
	if err := json.Unmarshal(params, &p); err != nil {
		log.L.Warnw("Malformed $/cancelRequest params", "err", err)
		return nil
	}
	idKey := requestIDKey(&p.Id)
	m.inflightMu.Lock()
	cancel, ok := m.inflight[idKey]
	m.inflightMu.Unlock()
	if !ok {
		log.L.Infow("$/cancelRequest for unknown request id (already completed?)", "id", idKey)
		return nil
	}
	log.L.Infow("Cancelling in-flight request", "id", idKey)
	cancel()
	return nil
}

func (m *Mux) registerInflight(idKey string, cancel context.CancelFunc) {
	m.inflightMu.Lock()
	// A misbehaving client could reuse an id while the prior request
	// is still in flight. Cancel the older one before stomping its
	// entry so its goroutine releases its context budget rather than
	// leaking until session shutdown.
	if old, ok := m.inflight[idKey]; ok {
		old()
	}
	m.inflight[idKey] = cancel
	m.inflightMu.Unlock()
}

func (m *Mux) unregisterInflight(idKey string) {
	m.inflightMu.Lock()
	delete(m.inflight, idKey)
	m.inflightMu.Unlock()
}

// requestIDKey canonicalizes a request id into a string key that
// survives JSON's number-vs-string ambiguity. We use the trimmed raw
// bytes; that way `1` and `"1"` are distinct keys (as they should be
// per JSON-RPC), and equal ids - whatever their representation -
// share a key.
func requestIDKey(id *json.RawMessage) string {
	if id == nil {
		return ""
	}
	return string(bytes.TrimSpace(*id))
}

func (m *Mux) write(msg any) (err error) {
	m.writeLock.Lock()
	defer m.writeLock.Unlock()
	return rpc.Encode(m.writer, msg)
}
