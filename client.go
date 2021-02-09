package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	binance "github.com/adshao/go-binance/futures"
)

const (
	BinanceClientType = "binance"

	_clientID = "swiper"
)

type Order struct {
	ID       string
	Symbol   string
	Type     string
	Quantity float64
	Side     string
	Price    string
}

type Client interface {
	ClosedOrders() (map[string]*Order, error)
	NewOrder(*Order) (*Order, error)
}

func NewClient(exchange string, key string, secret string) (Client, error) {
	if exchange != BinanceClientType {
		return nil, fmt.Errorf("exchange '%s' not supported, available are [%s]", exchange, BinanceClientType)
	}
	return &BinanceClient{
		binance.NewClient(key, secret),
	}, nil
}

type BinanceClient struct {
	client *binance.Client
}

func (b *BinanceClient) ClosedOrders() (map[string]*Order, error) {
	ordersRaw, err := b.client.NewListOrdersService().Limit(20).Do(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "read: connection reset by peer") {
			return b.ClosedOrders()
		}
		return nil, err
	}

	orders := make(map[string]*Order)
	for _, order := range ordersRaw {
		// ignore orders from phone
		// TODO: protect against iphone and web
		if strings.Contains(order.ClientOrderID, "android_") {
			continue
		}

		// ignore orders from swiper
		if strings.Contains(order.ClientOrderID, _clientID) {
			continue
		}

		quantity, err := strconv.ParseFloat(order.OrigQuantity, 64)
		if err != nil {
			return nil, err
		}

		orders[order.ClientOrderID] = &Order{
			ID:       order.ClientOrderID,
			Symbol:   order.Symbol,
			Type:     string(order.Type),
			Quantity: quantity,
			Side:     string(order.Side),
			Price:    order.AvgPrice,
		}
	}

	return orders, nil
}

func (b *BinanceClient) NewOrder(order *Order) (*Order, error) {
	newOrder, err := b.client.NewCreateOrderService().
		NewClientOrderID(_clientID).
		Symbol(order.Symbol).
		Type(binance.OrderType(order.Type)).
		Quantity(fmt.Sprintf("%v", order.Quantity)).
		Side(binance.SideType(order.Side)).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	order.ID = newOrder.ClientOrderID
	return order, nil
}

// TODO: update for kraken
// import(
// 	krakenapi "github.com/beldur/kraken-go-api-client"
// )
// type Kraken struct {
// 	api            *krakenapi.KrakenAPI
// 	previousOrders map[string]krakenapi.Order
// 	newOrders      map[string]krakenapi.Order
// }
//
// func NewKraken() *Kraken {
// 	api := krakenapi.New(krakenKey, krakenSecret)
//
// 	return &Kraken{api: api}
// }
//
// func (k *Kraken) Pull() error {
// 	orders, err := k.getClosedOrdes()
// 	if err != nil {
// 		return err
// 	}
//
// 	k.newOrders = k.findNewOrders(orders)
// 	k.previousOrders = orders
//
// 	return nil
// }
//
// func (k *Kraken) Process() error {
// 	args := make(map[string]string)
// 	args["leverage"] = "3:1" // to discriminate swiper trade, maybe use order.Description.Leverage
//
// 	for _, order := range k.newOrders {
// 		volStr := order.Volume
//
// 		vol, err := strconv.ParseFloat(order.Volume, 64)
// 		if err == nil {
// 			volStr = fmt.Sprintf("%f", ratio*vol)
// 		}
//
// 		newOrder, err := k.api.AddOrder(
// 			order.Description.AssetPair,
// 			order.Description.Type,
// 			order.Description.OrderType,
// 			volStr,
// 			args,
// 		)
// 		if err != nil {
// 			return err
// 		}
//
// 		log.Printf("copied: %s -> %s\n", order.Description.Order, newOrder.Description.Order)
// 	}
//
// 	return nil
// }
//
// func (k *Kraken) getClosedOrdes() (map[string]krakenapi.Order, error) {
// 	orders, err := k.api.ClosedOrders(nil)
// 	if err != nil {
// 		if strings.Contains(err.Error(), "(Response Content-Type is 'text/html', but should be 'application/json'.)") ||
// 			strings.Contains(err.Error(), "read: connection reset by peer") ||
// 			strings.Contains(err.Error(), "EService:Unavailable") {
// 			return k.getClosedOrdes()
// 		}
// 		return nil, err
// 	}
// 	return orders.Closed, nil
// }
//
// func (k *Kraken) findNewOrders(orders map[string]krakenapi.Order) map[string]krakenapi.Order {
// 	newOrders := make(map[string]krakenapi.Order)
//
// 	if len(k.previousOrders) == 0 {
// 		return newOrders
// 	}
//
// 	for index, order := range orders {
// 		if !knowedKrakenOrder(k.previousOrders[index]) && isNapbotsKrakenOrder(order) {
// 			newOrders[index] = order
// 		}
// 	}
// 	return newOrders
// }
//
// func isNapbotsKrakenOrder(o krakenapi.Order) bool {
// 	// napbots make trade with 4:1 leverage
// 	return o.Description.Leverage == "4:1"
// }
//
// func knowedKrakenOrder(o krakenapi.Order) bool {
// 	return o.Status != ""
// }
