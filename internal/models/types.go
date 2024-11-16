package models

import (
	"github.com/gorilla/websocket"
)

type KrakenClient struct {
	ApiKey     string
	ApiSecret  string
	WsConn     *websocket.Conn
	Prices     map[string]float64
	PrevPrices map[string]float64
	Balances   map[string]float64
}

type BalanceResponse struct {
	Error  []string          `json:"error"`
	Result map[string]string `json:"result"`
}

type AssetValue struct {
	Asset     string
	Balance   float64
	Price     float64
	PrevPrice float64
	USDValue  float64
}

var AssetMapping = map[string]string{
	"XETH": "ETH/USD",
	"SOL":  "SOL/USD",
	"XXBT": "XBT/USD",
	"ZUSD": "USD",
}
