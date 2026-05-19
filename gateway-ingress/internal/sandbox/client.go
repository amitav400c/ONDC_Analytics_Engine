package sandbox

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client connects to the Rust edge-sandbox over UDS for PII redaction.
// Falls back gracefully if the sandbox is unavailable.
type Client struct {
	socketPath string
	conn       *grpc.ClientConn
	connected  bool
	mu         sync.RWMutex
}

func NewClient(socketPath string) *Client {
	c := &Client{socketPath: socketPath}
	go c.connectLoop()
	return c
}

func (c *Client) connectLoop() {
	for {
		conn, err := grpc.NewClient(
			"unix://"+c.socketPath,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
				return net.DialTimeout("unix", c.socketPath, 2*time.Second)
			}),
		)
		if err != nil {
			log.Printf("sandbox connection pending (will retry): %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		c.mu.Lock()
		c.conn = conn
		c.connected = true
		c.mu.Unlock()
		log.Println("connected to edge-sandbox via UDS")
		return
	}
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Sanitize sends payload to the WASM sandbox for PII redaction.
// TODO: Replace with generated protobuf client once sandbox.proto is compiled
func (c *Client) Sanitize(ctx context.Context, payload string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.conn == nil {
		return payload, nil // Passthrough
	}

	// TODO: Use generated SandboxServiceClient here
	// For Phase 1 (dumb pipeline), we just pass through
	return payload, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
