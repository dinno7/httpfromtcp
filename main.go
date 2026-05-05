package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

var bytePool = sync.Pool{
	New: func() any {
		b := make([]byte, 8)
		return &b
	},
}

func main() {
	messagesFile, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatalln(err)
	}

	buf := bytePool.Get().(*[]byte)
	for {
		n, err := messagesFile.Read(*buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalln(err)
			return
		}
		fmt.Printf("read: %s\n", string((*buf)[:n]))
	}

	bytePool.Put(buf)
}
