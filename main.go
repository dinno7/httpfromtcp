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
		newLineIndex := bytes.Index(readBytes, []byte("\n"))
		if newLineIndex != -1 {
			line = append(line, readBytes[:newLineIndex]...)
			fmt.Printf("read: %s\n", string(line))

			line = readBytes[newLineIndex+1:]
		} else {
			line = append(line, readBytes...)
		}

	}
}
