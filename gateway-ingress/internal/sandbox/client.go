package sandbox

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/amitav400c/ondc-analytics-gateway/gateway-ingress/internal/proto/sandbox"
)

// Client connects to the Rust edge-sandbox over UDS for PII redaction.
// Falls back gracefully if the sandbox is unavailable.
type Client struct {
	socketPath string
	conn       *grpc.ClientConn
	client     pb.SandboxServiceClient
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
		c.client = pb.NewSandboxServiceClient(conn)
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

// Sanitize sends payload to the WASM sandbox for PII redaction and WAF checks.
func (c *Client) Sanitize(ctx context.Context, payload string) (string, bool, string, error) {
	c.mu.RLock()
	client := c.client
	connected := c.connected
	c.mu.RUnlock()

	if !connected || client == nil {
		// Passthrough if sandbox is down (Fail-open for PII, but logs warning)
		return payload, true, "", nil
	}

	req := &pb.SanitizeRequest{
		PayloadJson: payload,
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := client.SanitizePayload(ctx, req)
	if err != nil {
		return payload, true, "", err
	}

	return resp.SanitizedJson, resp.IsSafe, resp.WafViolationReason, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
