package main

import (
	"fmt"
	"log"
	"net"

	"github.com/dinno7/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go func() {
			fmt.Println("New connection accepted")
			req, err := request.RequestFromReader(conn)
			if err != nil {
				fmt.Printf("error in parsing request: %s", err)
			}
			fmt.Println(req.String())
		}()
	}
}
