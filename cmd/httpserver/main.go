package main

import (
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/dinno7/httpfromtcp/internal/headers"
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
			} else if req.RequestLine.RequestTarget == "/stream" {
				count := 100
				for count > 0 {
					bufLen := getRandomIntBetween(64, 512)
					buf := make([]byte, int(bufLen))
					n, err := cryptoRand.Read(buf)
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						return server.NewHandlerErrorInternalServerError([]byte(err.Error()))
					}
					_, _ = w.WriteChunkedBody(buf[:n])
					time.Sleep(time.Millisecond * time.Duration(rand.IntN(1000)))
					count--
				}
				_, _ = w.WriteChunkedBodyDone()
				return nil
			} else if req.RequestLine.RequestTarget == "/trailers" {
				result, err := http.Get("https://httpbin.org/html")
				if err != nil {
					return server.NewHandlerErrorInternalServerError([]byte(err.Error()))
				}
				w.Headers().Set("Trailer", "X-Content-SHA256, X-Content-Length")

				buf := make([]byte, 3)
				fullResponse := []byte{}
				for {
					n, err := result.Body.Read(buf)
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						return server.NewHandlerErrorInternalServerError([]byte(err.Error()))
					}
					fullResponse = append(fullResponse, buf[:n]...)
					w.WriteChunkedBody(buf[:n])
					time.Sleep(time.Millisecond * 10)
				}
				w.WriteChunkedBodyDone()
				hashed := sha256.Sum256(fullResponse)
				sha512.Sum512(fullResponse)
				trailers := headers.NewHeaders()
				trailers.Set("X-Content-SHA256", hex.EncodeToString(hashed[:]))
				contentLen := strconv.FormatInt(int64(len(fullResponse)), 10)
				trailers.Set("X-Content-Length", contentLen)
				w.WriteTrailers(trailers)
			} else if req.RequestLine.RequestTarget == "/stream/media" {
				file, err := os.Open("dinnoland.png")
				if err != nil {
					return server.NewHandlerErrorInternalServerError([]byte(err.Error()))
				}
				defer file.Close()
				w.Headers().Set("Content-Type", "image/png")
				for {
					buf := make([]byte, getRandomIntBetween(1024*1, 1024*100))
					n, err := file.Read(buf)
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						return server.NewHandlerErrorInternalServerError([]byte(err.Error()))
					}
					w.WriteChunkedBody(buf[:n])
					time.Sleep(time.Millisecond * time.Duration(rand.IntN(500)))
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

func getRandomIntBetween(min, max int) int {
	return rand.IntN(max-min+1) + min
}
