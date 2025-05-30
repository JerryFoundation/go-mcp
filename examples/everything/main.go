package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

type currentTimeReq struct {
	Timezone string `json:"timezone" description:"current time timezone"`
}

func main() {
	// new mcp server with stdio or sse transport
	srv, err := server.NewServer(
		getTransport(),
		server.WithServerInfo(protocol.Implementation{
			Name:    "current-time-v2-server",
			Version: "1.0.0",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	tool, err := protocol.NewTool("current time", "Get current time with timezone, Asia/Shanghai is default", currentTimeReq{})
	if err != nil {
		log.Fatalf("Failed to create tool: %v", err)
		return
	}

	// register tool and start mcp server
	srv.RegisterTool(tool, currentTime)
	// srv.RegisterResource()
	// srv.RegisterPrompt()
	// srv.RegisterResourceTemplate()

	errCh := make(chan error)
	go func() {
		errCh <- srv.Run()
	}()

	if err = signalWaiter(errCh); err != nil {
		log.Fatalf("signal waiter: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
}

func getTransport() (t transport.ServerTransport) {
	mode := ""
	port := ""
	flag.StringVar(&mode, "transport", "stdio", "The transport to use, should be \"stdio\" or \"sse\"")
	flag.StringVar(&port, "port", "8080", "sse server address")
	flag.Parse()

	if mode == "stdio" {
		log.Println("start current time mcp server with stdio transport")
		t = transport.NewStdioServerTransport()
	} else {
		addr := fmt.Sprintf("127.0.0.1:%s", port)
		log.Printf("start current time mcp server with sse transport, listen %s", addr)
		t, _ = transport.NewSSEServerTransport(addr)
	}

	return t
}

func currentTime(request *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	req := new(currentTimeReq)
	if err := protocol.VerifyAndUnmarshal(request.RawArguments, &req); err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation(req.Timezone)
	if err != nil {
		return nil, fmt.Errorf("parse timezone with error: %v", err)
	}
	text := fmt.Sprintf(`current time is %s`, time.Now().In(loc))

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			protocol.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}, nil
}

func signalWaiter(errCh chan error) error {
	signalToNotify := []os.Signal{syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM}
	if signal.Ignored(syscall.SIGHUP) {
		signalToNotify = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, signalToNotify...)

	select {
	case sig := <-signals:
		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			log.Printf("Received signal: %s\n", sig)
			// graceful shutdown
			return nil
		}
	case err := <-errCh:
		return err
	}

	return nil
}
