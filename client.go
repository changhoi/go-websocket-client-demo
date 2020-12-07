package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type Unit struct {
	AskPrice float64 `json:"ask_price"`
	BidPrice float64 `json:"bid_price"`
	AskSize  float64 `json:"ask_size"`
	BidSize  float64 `json:"bid_size"`
}

type Response struct {
	Type           string  `json:"type"`
	OrderBookUnits []Unit  `json:"orderbook_units,omitempty"`
	TradePrice     float64 `json:"trade_price,omitempty"`
	TradeVolume    float64 `json:"trade_volume,omitempty"`
	AskBid         string  `json:"ask_bid,omitempty"`
}

var addr = "api.upbit.com"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/websocket/v1"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var res Response
			err = json.Unmarshal(message, &res)
			if err != nil {
				log.Print("unmarshal error", err)
				return
			}

			switch res.Type {
			case "orderbook":
				unit := res.OrderBookUnits[0]
				orderBook := fmt.Sprintf("ask_price: %f, bid_price: %f, ask_size: %f, bid_size: %f ", unit.AskPrice, unit.BidPrice, unit.AskSize, unit.BidSize)
				log.Print(orderBook)
			case "trade":
				log.Printf("trade_price: %f, trade_volume: %f, ask_bid: %s", res.TradePrice, res.TradeVolume, res.AskBid)
			default:
				log.Print("err")
			}
		}
	}()

	err = c.WriteMessage(websocket.TextMessage, []byte("[{\"ticket\":\"UNIQUE_TICKET\"},{\"type\":\"trade\",\"codes\":[\"KRW-BTC\"]},{\"type\":\"orderbook\",\"codes\":[\"KRW-BTC\"]}]"))
	if err != nil {
		log.Println("write:", err)
		return
	}

	for {
		select {
		case <-done:
			return

		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
