package main

import (
	"log"
	"os"
	"strings"
	"time"

	krakenapi "github.com/beldur/kraken-go-api-client"
)

var (
	key    = os.Getenv("KRAKEN_API_KEY")
	secret = os.Getenv("KRAKEN_API_SECRET")

	previousOrders map[string]krakenapi.Order
	api            *krakenapi.KrakenAPI
)

func main() {
	api = krakenapi.New(key, secret)

	for {
		orders, err := getClosedOrdes()
		if err != nil {
			log.Fatalf("cannot get closed orders: %s", err.Error())
		}

		if err := processNewOrders(orders); err != nil {
			log.Fatalf("cannot process new orders: %s", err.Error())
		}

		previousOrders = orders
		time.Sleep(2100 * time.Millisecond)
	}
}

func getClosedOrdes() (map[string]krakenapi.Order, error) {
	orders, err := api.ClosedOrders(nil)
	if err != nil {
		if strings.Contains(err.Error(), "(Response Content-Type is 'text/html', but should be 'application/json'.)") ||
			strings.Contains(err.Error(), "read: connection reset by peer") ||
			strings.Contains(err.Error(), "EService:Unavailable") {
			return getClosedOrdes()
		}
		return nil, err
	}
	return orders.Closed, nil
}

func processNewOrders(orders map[string]krakenapi.Order) error {
	newOrders := checkNewOrders(orders)

	args := make(map[string]string)
	args["leverage"] = "2:1" // to discriminate swiper trade, maybe use order.Description.Leverage

	for _, order := range newOrders {
		if _, err := api.AddOrder(
			order.Description.AssetPair,
			order.Description.Type,
			order.Description.OrderType,
			order.Volume,
			args,
		); err != nil {
			return err
		}
		log.Println(order.Description.Order)
	}

	return nil
}

func checkNewOrders(orders map[string]krakenapi.Order) map[string]krakenapi.Order {
	newOrders := make(map[string]krakenapi.Order)

	if len(previousOrders) == 0 {
		return newOrders
	}

	for index, order := range orders {
		if !knowedOrder(previousOrders[index]) && isNapbotsOrder(order) {
			newOrders[index] = order
		}
	}
	return newOrders
}

func isNapbotsOrder(o krakenapi.Order) bool {
	// napbots make trade with 4:1 leverage
	return o.Description.Leverage == "4:1"
}

func knowedOrder(o krakenapi.Order) bool {
	return o.Status != ""
}
