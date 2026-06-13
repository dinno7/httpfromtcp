package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dinno7/httpfromtcp/internal/request"
	"github.com/dinno7/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.Method {
		case "GET":
			switch req.RequestLine.RequestTarget {
			case "/":
				w.Write([]byte("{\"age\": 25}"))
			case "/error":
				msg, _ := json.Marshal(map[string]any{
					"ok":      false,
					"message": "God Damn",
				})
				return server.NewHandlerErrorBadRequest(msg)
			default:
				return server.NewHandlerErrorNotFound([]byte("Not found!"))
			}
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
