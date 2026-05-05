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
	defer messagesFile.Close()

	for line := range getLineChannel(messagesFile) {
		fmt.Printf("read: %s\n", line)
	}
}

func getLineChannel(f io.ReadCloser) <-chan string {
	lineChan := make(chan string)
	go func() {
		defer f.Close()
		defer close(lineChan)

		line := []byte{}
		for {
			buf := make([]byte, 8)
			n, err := f.Read(buf)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Println(err)
				}

				// NOTE: Send to channel remain last line
				if len(line) > 0 {
					lineChan <- string(line)
				}
				break
			}
			readBytes := buf[:n]
			newLineIndex := bytes.Index(readBytes, []byte("\n"))
			if newLineIndex != -1 {
				line = append(line, readBytes[:newLineIndex]...)
				lineChan <- string(line) // NOTE: Send to whole line
				line = readBytes[newLineIndex+1:]
			} else {
				line = append(line, readBytes...)
			}
		}
	}()

	return lineChan
}
