package main

import (
	"log"
	"time"
)

var (
	client Client
)

func main() {
	client = NewBinance()
	// client = NewKraken()

	for {
		loop()
		time.Sleep(2100 * time.Millisecond)
	}
}

func loop() {
	if err := client.Pull(); err != nil {
		log.Printf("cannot get closed orders: %s\n", err.Error())
		return
	}

	if err := client.Process(); err != nil {
		log.Printf("cannot process new orders: %s\n", err.Error())
		return
	}
}
