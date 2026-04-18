package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/agentshield/agents/internal/service"
)

func main() {
	port := os.Getenv("AGENTS_PORT")
	if port == "" {
		port = "8090"
	}

	mux := http.NewServeMux()
	h := service.NewHandler()
	h.Register(mux)

	addr := ":" + port
	log.Printf("agents service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
