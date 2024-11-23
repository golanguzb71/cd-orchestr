package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "jprq-event/protos/pb"
)

type TunnelConnection struct {
	Port         string
	WsConn       *websocket.Conn
	ResponseChan chan []byte
	mu           sync.Mutex
}

type EventServer struct {
	pb.UnimplementedEventServiceServer
	rdb         *redis.Client
	connections map[string]*TunnelConnection
	mu          sync.RWMutex
}

func NewEventServer(rdb *redis.Client) *EventServer {
	return &EventServer{
		rdb:         rdb,
		connections: make(map[string]*TunnelConnection),
	}
}

func (s *EventServer) HandleRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	s.mu.RLock()
	conn, exists := s.connections[req.Domain]
	s.mu.RUnlock()

	if !exists || conn.WsConn == nil {
		return nil, status.Errorf(codes.NotFound, "no active connection for domain")
	}

	// Prepare full request payload
	wsRequest := map[string]interface{}{
		"method":  req.Method,
		"path":    req.Path,
		"headers": req.Headers,
		"body":    req.Body,
	}

	// Synchronize WebSocket access
	conn.mu.Lock()
	defer conn.mu.Unlock()

	// Reset response channel
	conn.ResponseChan = make(chan []byte, 1)

	// Send request through WebSocket
	if err := conn.WsConn.WriteJSON(wsRequest); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write to websocket: %v", err)
	}

	// Wait for response with timeout
	select {
	case <-ctx.Done():
		return nil, status.Errorf(codes.DeadlineExceeded, "timeout while reading response")
	case message := <-conn.ResponseChan:
		// Parse WebSocket response
		var response struct {
			StatusCode int32             `json:"status_code"`
			Headers    map[string]string `json:"headers"`
			Body       json.RawMessage   `json:"body"`
		}

		if err := json.Unmarshal(message, &response); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse response: %v", err)
		}

		return &pb.Response{
			StatusCode: response.StatusCode,
			Headers:    response.Headers,
			Body:       map[string]string{"message": string(response.Body)},
		}, nil
	}
}

func (s *EventServer) RegisterConnection(domain string, conn *TunnelConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[domain] = conn
}

func (s *EventServer) UnregisterConnection(domain string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connections, domain)
}

type WebSocketHandler struct {
	rdb         *redis.Client
	eventServer *EventServer
}

func NewWebSocketHandler(rdb *redis.Client, eventServer *EventServer) *WebSocketHandler {
	return &WebSocketHandler{
		rdb:         rdb,
		eventServer: eventServer,
	}
}

func (h *WebSocketHandler) HandleTunnel(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	domain := r.URL.Query().Get("domain")
	port := r.URL.Query().Get("port")

	if domain == "" || port == "" {
		sendErrorResponse(conn, "Domain and Port are required", http.StatusBadRequest)
		return
	}

	// Store port in Redis with a 24-hour expiration
	ctx := context.Background()
	err = h.rdb.Set(ctx, domain, port, 24*time.Hour).Err()
	if err != nil {
		sendErrorResponse(conn, "Failed to store port", http.StatusInternalServerError)
		return
	}

	tunnelConn := &TunnelConnection{
		Port:         port,
		WsConn:       conn,
		ResponseChan: make(chan []byte, 1),
	}
	h.eventServer.RegisterConnection(domain, tunnelConn)
	defer h.eventServer.UnregisterConnection(domain)

	// Keep WebSocket connection open and handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("WebSocket closed unexpectedly:", err)
			}
			break
		}

		// Send received message to the response channel
		tunnelConn.ResponseChan <- message
		fmt.Println("Received message in HandleTunnel:", string(message))
	}
}

func sendErrorResponse(conn *websocket.Conn, message string, statusCode int) {
	response := map[string]interface{}{
		"status_code": statusCode,
		"headers":     map[string]string{"Content-Type": "application/json"},
		"body":        message,
	}
	conn.WriteJSON(response)
}
