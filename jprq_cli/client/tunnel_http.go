package main

import (
	"context"
	"fmt"
	"log"
	urlpkg "net/url"
	"os/user"
)

type HTTPTunnel struct {
	Host    string `bson:"host"`
	Token   string `bson:"token"`
	Warning string `bson:"warning"`
	Error   string `bson:"error"`
}

func openHTTPTunnel(port int, subdomain string, ctx context.Context) {
	if subdomain == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatalf("Please specify -subdomain")
		}
		subdomain = u.Username
	}
	query := fmt.Sprintf("port=%d&username=%s&version=%s", port, subdomain, version)
	url := urlpkg.URL{Scheme: "wss", Host: httpBaseHost, Path: "/_ws/", RawQuery: query}

}
