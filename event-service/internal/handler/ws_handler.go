package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "jprq-event/protos/pb"
	"net/http"
	"strings"
	"time"
)

type TunnelConnection struct {
	Port   string
	WsConn *websocket.Conn
}

type EventServer struct {
	pb.UnimplementedEventServiceServer
	rdb         *redis.Client
	connections map[string]*TunnelConnection
}

func NewEventServer(rdb *redis.Client) *EventServer {
	return &EventServer{
		rdb:         rdb,
		connections: make(map[string]*TunnelConnection),
	}
}

func (s *EventServer) HandleRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	// Get the port for the given domain
	port, err := s.rdb.Get(ctx, req.Domain).Result()
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "domain not found: %v", err)
	}

	// Get the connection for the domain
	conn, exists := s.connections[req.Domain]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "no active connection for domain")
	}

	// Forward the request through WebSocket
	wsRequest := map[string]interface{}{
		"method":  req.Method,
		"path":    req.Path,
		"headers": req.Headers,
		"body":    req.Body,
		"port":    port,
	}
	if err := conn.WsConn.WriteJSON(wsRequest); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write to websocket: %v", err)
	}

	// Read the response from WebSocket
	_, message, err := conn.WsConn.ReadMessage()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read from websocket: %v", err)
	}

	// Return the response as is (no additional processing)
	var response struct {
		StatusCode int               `json:"status_code"`
		Headers    map[string]string `json:"headers"`
		Body       string            `json:"body"`
	}

	if err := json.Unmarshal(message, &response); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse response: %v", err)
	}

	// Return the response directly
	return &pb.Response{
		StatusCode: int32(response.StatusCode),
		Headers:    response.Headers,
		Body:       map[string]string{"message": response.Body},
	}, nil
}

func (s *EventServer) RegisterConnection(domain string, conn *TunnelConnection) {
	s.connections[domain] = conn
}

func (s *EventServer) UnregisterConnection(domain string) {
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

	tunnelConn := &TunnelConnection{
		Port:   port,
		WsConn: conn,
	}
	h.eventServer.RegisterConnection(domain, tunnelConn)
	defer h.eventServer.UnregisterConnection(domain)

	h.rdb.Set(context.TODO(), domain, port, 24*time.Hour)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("WebSocket closed unexpectedly:", err)
			}
			break
		}

		if messageType != websocket.TextMessage {
			continue
		}

		// Just forward the incoming WebSocket message as is
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			fmt.Println("Error forwarding message:", err)
			break
		}
	}
}

func isValidMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	method = strings.ToUpper(method)
	for _, valid := range validMethods {
		if valid == method {
			return true
		}
	}
	return false
}

func sendErrorResponse(conn *websocket.Conn, message string, statusCode int) {
	response := map[string]interface{}{
		"status_code": statusCode,
		"headers":     map[string]string{"Content-Type": "application/json"},
		"body":        message,
	}
	conn.WriteJSON(response)
}

func forwardMessageAsIs(conn *websocket.Conn, message []byte) {
	conn.WriteMessage(websocket.TextMessage, message)
}
