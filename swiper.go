package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	binance "github.com/adshao/go-binance/futures"
)

var (
	binanceFetcherKey    = os.Getenv("BINANCE_API_KEY")
	binanceFetcherSecret = os.Getenv("BINANCE_API_SECRET")
	binanceCopycatKey    = os.Getenv("BINANCE_API_COPYCAT_KEY")
	binanceCopycatSecret = os.Getenv("BINANCE_API_COPYCAT_SECRET")

	ratio float64 = 2
)

type Swiper interface {
	Run() error
}

type Binance struct {
	mu sync.Mutex

	apiFetcher *binance.Client
	apiCopycat *binance.Client

	lastTickOrders map[string]*binance.Order
	copiedOrderID  map[string]bool
}

func NewBinanceSwiper() *Binance {
	if binanceCopycatKey == "" {
		binanceCopycatKey = binanceFetcherKey
		binanceCopycatSecret = binanceFetcherSecret
	}

	apiFetcher := binance.NewClient(binanceFetcherKey, binanceFetcherSecret)
	apiCopycat := binance.NewClient(binanceCopycatKey, binanceCopycatSecret)
	return &Binance{
		apiFetcher:    apiFetcher,
		apiCopycat:    apiCopycat,
		copiedOrderID: make(map[string]bool),
	}
}

func (b *Binance) Run() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	orders, err := b.getClosedOrders()
	if err != nil {
		return fmt.Errorf("cannot get closed orders: %s", err.Error())
	}

	newOrders := b.findNewOrders(orders)
	b.lastTickOrders = orders

	if err := b.copy(newOrders); err != nil {
		return fmt.Errorf("cannot process new orders: %s", err.Error())
	}

	return nil
}

func (b *Binance) getClosedOrders() (map[string]*binance.Order, error) {
	ordersRaw, err := b.apiFetcher.NewListOrdersService().Limit(20).Do(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "read: connection reset by peer") {
			return b.getClosedOrders()
		}
		return nil, err
	}

	orders := make(map[string]*binance.Order)
	for _, order := range ordersRaw {
		orders[order.ClientOrderID] = order
	}

	return orders, nil
}

func (b *Binance) findNewOrders(orders map[string]*binance.Order) map[string]*binance.Order {
	newOrders := make(map[string]*binance.Order)

	if len(b.lastTickOrders) == 0 {
		return newOrders
	}

	for index, order := range orders {
		// dont copy orders made from android
		// TODO: add iphone protection
		if b.lastTickOrders[index] == nil && !strings.Contains(index, "android_") {
			if b.copiedOrderID[index] {
				delete(b.copiedOrderID, index)
				continue
			}
			newOrders[index] = order
		}
	}
	return newOrders
}

func (b *Binance) copy(newOrders map[string]*binance.Order) error {
	b.apiCopycat.NewCreateOrderService().Type(binance.OrderTypeMarket)

	// TODO: parallelize this with chan
	for _, order := range newOrders {
		quantity, err := strconv.ParseFloat(order.OrigQuantity, 64)
		if err != nil {
			return err
		}
		quantity = quantity * ratio
		quantityStr := fmt.Sprintf("%v", quantity)
		newOrder, err := b.apiCopycat.NewCreateOrderService().Symbol(order.Symbol).Type(order.Type).Quantity(quantityStr).Side(order.Side).Do(context.Background())
		if err != nil {
			return err
		}

		// TODO: use order.Fills to calculate executed price (now its just "MARKET")
		// wait for chan to be implemented or it will add processing time
		// and time is money
		log.Printf("copied: %s %s %s %s @ %s \n\t-> %s %s %s %s @ %s\n",
			order.ClientOrderID, order.Side, order.OrigQuantity, order.Symbol, order.Type,
			newOrder.ClientOrderID, newOrder.Side, quantityStr, newOrder.Symbol, newOrder.Type,
		)

		b.copiedOrderID[newOrder.ClientOrderID] = true
	}

	return nil
}
