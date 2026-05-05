package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:46069")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	conn.Write([]byte("hi there"))
	buf := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buf)
	fmt.Println("Got: ", string(buf[:n]))
}
