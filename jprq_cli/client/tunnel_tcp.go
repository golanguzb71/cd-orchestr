package main

import "context"

type TCPTunnel struct {
	PublicServerPort  int `json:"public_server_port"`
	PrivateServerPort int `json:"private_server_port"`
}

type ConnectionRequest struct {
	Flag int `json:"public_client_port"`
}

func openTCPTunnel(port int, ctx context.Context) {

}
