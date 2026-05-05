package main

import (
	"fmt"
	"log"
	"net"
)

var incommingBufferSize = 10 * 1024

func main() {
	addr, _ := net.ResolveUDPAddr("udp", ":46069")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, incommingBufferSize)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Got: ", string(buf[:n]))
		conn.WriteToUDP([]byte("bye there"), addr)
	}
}
