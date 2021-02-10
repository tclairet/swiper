package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	fetcherExchange = os.Getenv("FETCHER_EXCHANGE")
	fetcherKey      = os.Getenv("FETCHER_API_KEY")
	fetcherSecret   = os.Getenv("FETCHER_API_SECRET")
	ratioStr        = os.Getenv("RATIO")
)

func main() {
	client, err := NewClient(fetcherExchange, fetcherKey, fetcherSecret)
	if err != nil {
		log.Fatal(err)
	}

	config := &DefaultConfig
	if ratioStr != "" {
		r, err := strconv.ParseFloat(ratioStr, 64)
		if err != nil {
			log.Fatal(err)
		}
		config.Ratio = r
	}
	swiper := NewSwiper(client, config)

	for {
		res, err := swiper.Run()
		if err != nil {
			if strings.Contains(err.Error(), "code=-1021") || strings.Contains(err.Error(), "code=-1000") {
				continue
			}
			log.Printf("swiper: %s\n", err.Error())
		}

		for order, copiedOrders := range res {
			// TODO: print market price (now it's just "MARKET" given by Type)
			log.Printf("copied: %s %s %f %s @ %s:%s\n",
				order.ID, order.Side, order.Quantity, order.Symbol, order.Type, order.Price,
			)
			for _, copiedOrder := range copiedOrders {
				log.Printf("\t-> %s %s %f %s @ %s:%s\n",
					copiedOrder.ID, copiedOrder.Side, copiedOrder.Quantity, copiedOrder.Symbol, copiedOrder.Type, copiedOrder.Price,
				)
			}
		}

		time.Sleep(2100 * time.Millisecond)
	}
}
