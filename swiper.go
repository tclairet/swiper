package main

import (
	"fmt"
	"sync"
)

var (
	ratio         float64 = 5
	DefaultConfig         = Config{
		CopyOnFetcher:  true,
		CopycatClients: nil,
	}
)

type Config struct {
	CopyOnFetcher  bool
	CopycatClients []Client
}

type Swiper struct {
	mu sync.Mutex

	client         Client
	config         Config
	lastTickOrders map[string]*Order
}

func NewSwiper(client Client, config *Config) *Swiper {
	if config == nil {
		config = &DefaultConfig
	}

	s := &Swiper{
		client: client,
		config: *config,
	}

	return s
}

func (s *Swiper) Run() (map[*Order][]*Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	orders, err := s.client.ClosedOrders()
	if err != nil {
		return nil, fmt.Errorf("cannot get closed orders: %s", err.Error())
	}

	newOrders := s.findNewOrders(orders)
	s.lastTickOrders = orders

	result, err := s.copy(newOrders)
	if err != nil {
		return nil, fmt.Errorf("cannot process new orders: %s", err.Error())
	}

	return result, nil
}

func (s *Swiper) findNewOrders(orders map[string]*Order) []*Order {
	var newOrders []*Order

	if len(s.lastTickOrders) == 0 {
		return newOrders
	}

	for index, order := range orders {
		if s.lastTickOrders[index] == nil {
			newOrders = append(newOrders, order)
		}
	}
	return newOrders
}

func makeCopyOrder(order *Order) *Order {
	return &Order{
		Symbol:   order.Symbol,
		Type:     order.Type,
		Quantity: order.Quantity * ratio,
		Side:     order.Side,
	}
}

func (s *Swiper) copy(orders []*Order) (map[*Order][]*Order, error) {
	res := make(map[*Order][]*Order, len(orders))

	// TODO: parallelize this with chan
	for _, order := range orders {
		copyOrders, err := s.sendOrder(makeCopyOrder(order))
		if err != nil {
			return nil, err
		}

		res[order] = copyOrders
	}

	return res, nil
}

func (s *Swiper) sendOrder(order *Order) ([]*Order, error) {
	var copyOrders []*Order

	if s.config.CopyOnFetcher {
		copyOrder, err := s.client.NewOrder(order)
		if err != nil {
			return nil, err
		}
		copyOrders = append(copyOrders, copyOrder)
	}

	for _, client := range s.config.CopycatClients {
		copyOrder, err := client.NewOrder(order)
		if err != nil {
			return nil, err
		}
		copyOrders = append(copyOrders, copyOrder)
	}

	return copyOrders, nil
}
