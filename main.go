package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/divyam234/installer/handler"
)

func main() {
	c := handler.DefaultConfig

	if c.ForceUser != "" {
		log.Printf("locked user to '%s'", c.ForceUser)
	}
	if c.ForceRepo != "" {
		log.Printf("locked repo to '%s'", c.ForceRepo)
	}

	lh := &handler.Handler{Config: c}
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s...", addr)
	if err := http.Serve(l, lh); err != nil {
		log.Fatal(err)
	}
	log.Print("exiting")
}
