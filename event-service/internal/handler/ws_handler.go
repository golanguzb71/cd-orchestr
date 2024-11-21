package handlers

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "jprq-event/protos/pb"
	"log"
	"net/http"
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
	// Get the port for the given domain from Redis
	port, err := s.rdb.Get(ctx, req.Domain).Result()
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "domain not found: %v", err)
	}

	// Retrieve the connection for the domain
	conn, exists := s.connections[req.Domain]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "no active connection for domain")
	}

	fmt.Printf("Forwarding request for domain %s (port: %s)\n", req.Domain, port)

	// Prepare the request data to be forwarded over WebSocket
	wsRequest := map[string]interface{}{
		"method":  req.Method,
		"path":    req.Path,
		"headers": req.Headers,
		"body":    req.Body,
	}

	// Send the request over the WebSocket connection
	if err := conn.WsConn.WriteJSON(wsRequest); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write to websocket: %v", err)
	}

	// Receive the response from WebSocket
	var wsResponse map[string]interface{}
	if err := conn.WsConn.ReadJSON(&wsResponse); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read from websocket: %v", err)
	}

	// Prepare the response from WebSocket data
	response := &pb.Response{
		StatusCode: int32(wsResponse["status_code"].(float64)),
		Headers:    make(map[string]string),
		Body:       wsResponse["body"].([]byte),
	}

	if headers, ok := wsResponse["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			response.Headers[k] = fmt.Sprint(v)
		}
	}

	return response, nil
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
	// Upgrade HTTP request to WebSocket connection
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		http.Error(w, "Failed to upgrade to WebSocket: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Get domain and port from the query parameters
	domain := r.URL.Query().Get("domain")
	port := r.URL.Query().Get("port")

	if domain == "" || port == "" {
		// If the domain or port is missing, send a WebSocket-specific error message
		log.Println("Domain and Port are required")
		conn.WriteMessage(websocket.TextMessage, []byte("Domain and Port are required"))
		return
	}

	// Register the WebSocket connection for the domain
	tunnelConn := &TunnelConnection{
		Port:   port,
		WsConn: conn,
	}
	h.eventServer.RegisterConnection(domain, tunnelConn)
	h.rdb.Set(context.TODO(), domain, port, 100000*time.Second)
	// Handle WebSocket communication
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			break
		}

		// Process the received message
		fmt.Printf("Received message: %v\n", msg)

		// Example: You can send a response back via WebSocket if needed
		// e.g., conn.WriteJSON(response)
	}

	// Unregister the connection once done
	h.eventServer.UnregisterConnection(domain)
}
