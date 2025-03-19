package server

import (
	"fmt"
	"sync/atomic"

	"go-mcp/pkg"
	"go-mcp/protocol"
	"go-mcp/transport"
)

type Server struct {
	transport transport.Transport

	reqID2respChan map[int]chan *protocol.JSONRPCResponse

	notifyMethod2handler map[protocol.Method]func(notifyParam interface{})

	requestID atomic.Int64

	logger pkg.Logger
}

func NewServer(t transport.Transport, opts ...Option) (*Server, error) {
	server := &Server{
		transport: t,
		logger:    &pkg.Log{},
	}
	t.SetReceiver(server)

	for _, opt := range opts {
		opt(server)
	}

	return server, nil
}

func (server *Server) Start() error {
	if err := server.transport.Start(); err != nil {
		return fmt.Errorf("init mcp server transpor start fail: %w", err)
	}
	return nil
}

type Option func(*Server)

func WithNotifyHandler(notifyMethod2handler map[protocol.Method]func(notifyParam interface{})) Option {
	return func(s *Server) {
		s.notifyMethod2handler = notifyMethod2handler
	}
}

func WithLogger(logger pkg.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

func (server *Server) Shutdown() error {
	// TODO 还有一些其他处理操作也可以放在这里
	// server.transport.Close()
	return nil
}
