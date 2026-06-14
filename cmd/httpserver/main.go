package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dinno7/httpfromtcp/internal/request"
	"github.com/dinno7/httpfromtcp/internal/response"
	"github.com/dinno7/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w *response.Response, req *request.Request) *server.HandlerError {
		if req.RequestLine.Method == "GET" {
			if req.RequestLine.RequestTarget == "/" {
				w.Headers().Set("Content-Type", "text/html")
				w.Write(
					[]byte(
						"<html> <head> <title>200 OK</title> </head> <body> <h1>Success!</h1> <p>Your request was an absolute banger.</p> </body> </html>",
					),
				)

			} else if req.RequestLine.RequestTarget == "/error" {
				msg, _ := json.Marshal(map[string]any{
					"ok":      false,
					"message": "God Damn",
				})
				w.Headers().Set("Content-Type", "application/json")
				return server.NewHandlerErrorBadRequest(msg)
			} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbun") {
				targetUrl := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbun/")
				buf := make([]byte, 1024)
				for {
					result, err := http.Get(fmt.Sprintf("https://httpbun.com/%s", targetUrl))
					if err != nil {
						return server.NewHandlerErrorInternalServerError([]byte(err.Error()))
					}
					reader := result.Body
					n, err := reader.Read(buf)
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						return server.NewHandlerErrorInternalServerError([]byte(err.Error()))
					}
					w.WriteChunkedBody(buf[:n])
				}
				w.WriteChunkedBodyDone()
			} else {
				return server.NewHandlerErrorNotFound(
					[]byte(
						"<html> <head> <title>404 Not Found</title> </head> <body> <h1>Not Found!</h1> <p>No anything found for your path.</p> </body> </html>",
					),
				)
			}
		}

		return nil
	}

	server, err := server.Serve(
		port,
		handler,
	)
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
