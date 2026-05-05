package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	messagesFile, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		buf := make([]byte, 8)
		n, err := messagesFile.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalln(err)
			return
		}
		fmt.Printf("read: %s\n", string(buf[:n]))
	}
}
