package main_test

import (
	"testing"

	swiper "github.com/tclairet/swiper"
)

func TestSwiper(t *testing.T) {
	c := &mokeClient{
		closedOrders: map[string]*swiper.Order{
			"1": &swiper.Order{
				ID:       "1",
				Symbol:   "TESTUSDT",
				Type:     "MARKET",
				Quantity: float64(1),
				Side:     "BUY",
			},
		},
	}

	t.Run("First Run", func(t *testing.T) {
		s := swiper.NewSwiper(c, nil)
		s.Run()
		c.closedOrders["2"] = &swiper.Order{
			ID:       "2",
			Symbol:   "TESTUSDT",
			Type:     "MARKET",
			Quantity: float64(5),
			Side:     "BUY",
		}
		res, err := s.Run()
		if err != nil {
			t.Fatal(err)
		}

		if got, want := len(res), 1; got != want {
			t.Fatalf("got %v, want %v", got, want)
		}

		for order, copiedOrders := range res {
			if got, want := order.ID, "2"; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := len(copiedOrders), 1; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	})
}

type mokeClient struct {
	closedOrders map[string]*swiper.Order
	err          error
}

func (c *mokeClient) ClosedOrders() (map[string]*swiper.Order, error) {
	orders := make(map[string]*swiper.Order, len(c.closedOrders))
	for k, v := range c.closedOrders {
		orders[k] = v
	}
	return orders, c.err
}

func (c *mokeClient) NewOrder(order *swiper.Order) (*swiper.Order, error) {
	order.ID = "toto"
	return order, c.err
}
