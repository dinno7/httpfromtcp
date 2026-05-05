package main

import (
	"bytes"
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

	line := []byte{}
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
		readBytes := buf[:n]
		if bytes.Contains(readBytes, []byte("\n")) {
			splitted := bytes.Split(readBytes, []byte("\n"))
			line = append(line, splitted[0]...)
			fmt.Printf("read: %s\n", string(line))

			line = splitted[1]
		} else {
			line = append(line, readBytes...)
		}

	}
}
