package main

import (
	"log"
	"os"
	"time"
)

var (
	fetcherExchange = os.Getenv("FETCHER_EXCHANGE")
	fetcherKey      = os.Getenv("FETCHER_API_KEY")
	fetcherSecret   = os.Getenv("FETCHER_API_SECRET")
)

func main() {
	client, err := NewClient(fetcherExchange, fetcherKey, fetcherSecret)
	if err != nil {
		log.Fatal(err)
	}

	config := &DefaultConfig
	swiper := NewSwiper(client, config)

	for {
		res, err := swiper.Run()
		if err != nil {
			log.Printf("swiper: %s\n", err.Error())
		}

		for order, copiedOrders := range res {
			// TODO: print market price (now it's just "MARKET" given by Type)
			log.Printf("copied: %s %s %f %s @ %s\n",
				order.ID, order.Side, order.Quantity, order.Symbol, order.Type,
			)
			for _, copiedOrder := range copiedOrders {
				log.Printf("\t-> %s %s %f %s @ %s\n",
					copiedOrder.ID, copiedOrder.Side, copiedOrder.Quantity, copiedOrder.Symbol, copiedOrder.Type,
				)
			}
		}

		time.Sleep(2100 * time.Millisecond)
	}
}
