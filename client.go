package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	binance "github.com/adshao/go-binance/futures"
	krakenapi "github.com/beldur/kraken-go-api-client"
)

var (
	krakenKey    = os.Getenv("KRAKEN_API_KEY")
	krakenSecret = os.Getenv("KRAKEN_API_SECRET")

	binanceKey    = os.Getenv("BINANCE_API_KEY")
	binanceSecret = os.Getenv("BINANCE_API_SECRET")

	ratio float64 = 1
)

type Client interface {
	Pull() error
	Process() error
}

type Kraken struct {
	api            *krakenapi.KrakenAPI
	previousOrders map[string]krakenapi.Order
	newOrders      map[string]krakenapi.Order
}

func NewKraken() *Kraken {
	api := krakenapi.New(krakenKey, krakenSecret)

	return &Kraken{api: api}
}

func (k *Kraken) Pull() error {
	orders, err := k.getClosedOrdes()
	if err != nil {
		return err
	}

	k.newOrders = k.findNewOrders(orders)
	k.previousOrders = orders

	return nil
}

func (k *Kraken) Process() error {
	args := make(map[string]string)
	args["leverage"] = "3:1" // to discriminate swiper trade, maybe use order.Description.Leverage

	for _, order := range k.newOrders {
		volStr := order.Volume

		vol, err := strconv.ParseFloat(order.Volume, 64)
		if err == nil {
			volStr = fmt.Sprintf("%f", ratio*vol)
		}

		newOrder, err := k.api.AddOrder(
			order.Description.AssetPair,
			order.Description.Type,
			order.Description.OrderType,
			volStr,
			args,
		)
		if err != nil {
			return err
		}

		log.Printf("copied: %s -> %s\n", order.Description.Order, newOrder.Description.Order)
	}

	return nil
}

func (k *Kraken) getClosedOrdes() (map[string]krakenapi.Order, error) {
	orders, err := k.api.ClosedOrders(nil)
	if err != nil {
		if strings.Contains(err.Error(), "(Response Content-Type is 'text/html', but should be 'application/json'.)") ||
			strings.Contains(err.Error(), "read: connection reset by peer") ||
			strings.Contains(err.Error(), "EService:Unavailable") {
			return k.getClosedOrdes()
		}
		return nil, err
	}
	return orders.Closed, nil
}

func (k *Kraken) findNewOrders(orders map[string]krakenapi.Order) map[string]krakenapi.Order {
	newOrders := make(map[string]krakenapi.Order)

	if len(k.previousOrders) == 0 {
		return newOrders
	}

	for index, order := range orders {
		if !knowedKrakenOrder(k.previousOrders[index]) && isNapbotsKrakenOrder(order) {
			newOrders[index] = order
		}
	}
	return newOrders
}

func isNapbotsKrakenOrder(o krakenapi.Order) bool {
	// napbots make trade with 4:1 leverage
	return o.Description.Leverage == "4:1"
}

func knowedKrakenOrder(o krakenapi.Order) bool {
	return o.Status != ""
}

type Binance struct {
	api            *binance.Client
	lastOrders     map[string]*binance.Order
	previousOrders map[string]*binance.Order
	newOrders      map[string]*binance.Order
}

func NewBinance() *Binance {
	api := binance.NewClient(binanceKey, binanceSecret)
	return &Binance{api: api}
}

func (b *Binance) getClosedOrdes() (map[string]*binance.Order, error) {
	ordersRaw, err := b.api.NewListOrdersService().Limit(20).Do(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "read: connection reset by peer") {
			return b.getClosedOrdes()
		}
		return nil, err
	}

	orders := make(map[string]*binance.Order)
	for _, order := range ordersRaw {
		orders[order.ClientOrderID] = order
	}

	return orders, nil
}

func (b *Binance) Pull() error {
	orders, err := b.getClosedOrdes()
	if err != nil {
		return err
	}

	b.newOrders = b.findNewOrders(orders)
	b.lastOrders = orders

	return nil
}

func (b *Binance) findNewOrders(orders map[string]*binance.Order) map[string]*binance.Order {
	newOrders := make(map[string]*binance.Order)

	if len(b.previousOrders) == 0 {
		return newOrders
	}

	for index, order := range orders {
		if b.previousOrders[index] == nil {
			newOrders[index] = order
		}
	}
	return newOrders
}

func (b *Binance) Process() error {
	b.previousOrders = b.lastOrders

	b.api.NewCreateOrderService().Type(binance.OrderTypeMarket)

	for _, order := range b.newOrders {
		newOrder, err := b.api.NewCreateOrderService().Symbol(order.Symbol).Type(order.Type).Quantity(order.ExecutedQuantity).Side(order.Side).Do(context.Background())
		if err != nil {
			return err
		}

		log.Printf("copied: %s %s %s @ %s -> %s %s %s @ %s\n",
			order.Side, order.ExecutedQuantity, order.Symbol, order.Type,
			newOrder.Side, newOrder.OrigQuantity, newOrder.Symbol, newOrder.Type,
		)
		b.previousOrders[newOrder.ClientOrderID] = &binance.Order{}
	}

	return nil
}
