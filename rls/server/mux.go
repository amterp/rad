package server

import (
	"bufio"
	"encoding/json"
	"io"
	"rls/com"
	"rls/log"
	"rls/lsp"
	"rls/rpc"
	"sync"
)

type NotificationHandler func(params json.RawMessage) (err error)
type RequestHandler func(params json.RawMessage) (result any, err error)

type Mux struct {
	reader               *bufio.Reader
	writer               *bufio.Writer
	notificationHandlers map[string]NotificationHandler
	requestHandlers      map[string]RequestHandler
	writeLock            *sync.Mutex // chan struct{} ?
}

func NewMux(r io.Reader, w io.Writer) *Mux {
	mux := Mux{
		reader:               bufio.NewReader(r),
		writer:               bufio.NewWriter(w),
		notificationHandlers: make(map[string]NotificationHandler),
		requestHandlers:      make(map[string]RequestHandler),
		writeLock:            &sync.Mutex{},
	}
	return &mux
}

func (m *Mux) AddNotificationHandler(method string, handler NotificationHandler) {
	m.notificationHandlers[method] = handler
}

func (m *Mux) AddRequestHandler(method string, handler RequestHandler) {
	m.requestHandlers[method] = handler
}

func (m *Mux) Notify(notification any) (err error) {
	return m.write(notification)
}

func (m *Mux) Init() (err error) {
	log.L.Info("Initializing mux, awaiting initialize msg...")
	for {
		msg, err := rpc.Decode(m.reader)
		if err != nil {
			return err
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
				return err
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
	for {
		msg, err := rpc.Decode(m.reader)
		if err != nil {
			return err
		}
		err = m.handleMessage(msg)
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
	err = handler(*notification.Params)
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

	result, err := handler(*request.Params)
	if err != nil {
		log.L.Errorf("Error handling request %q: %v", request.Method, err)
		err = m.write(lsp.NewResponseError(request.Id, err))
		if err != nil {
			log.L.Errorf("Failed to write error response: %v", err)
		}
		return
	}
	log.L.Infow("Sending result", "result", com.FlatStr(result))

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

func (m *Mux) write(msg any) (err error) {
	m.writeLock.Lock()
	defer m.writeLock.Unlock()
	return rpc.Encode(m.writer, msg)
}
