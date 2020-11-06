package main

import (
	"log"
	"time"
)

var (
	swiper Swiper
)

func main() {
	swiper = NewBinanceSwiper()
	// client = NewKraken()

	for {
		if err := swiper.Run(); err != nil {
			log.Printf("swiper: %s\n", err.Error())
		}
		time.Sleep(2100 * time.Millisecond)
	}
}
