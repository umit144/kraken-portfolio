package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/umit144/kraken-portfolio/internal/config"
	"github.com/umit144/kraken-portfolio/internal/models"
	"github.com/umit144/kraken-portfolio/pkg/utils"

	"github.com/gorilla/websocket"
)

type Client struct {
	Config     *config.Config
	WsConn     *websocket.Conn
	Prices     map[string]float64
	PrevPrices map[string]float64
	Balances   map[string]float64
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		Config:     cfg,
		Prices:     make(map[string]float64),
		PrevPrices: make(map[string]float64),
		Balances:   make(map[string]float64),
	}
}

func (c *Client) GenerateSignature(urlPath, postData, nonce string) string {
	sha256Sum := sha256.Sum256([]byte(nonce + postData))
	decodedSecret, _ := base64.StdEncoding.DecodeString(c.Config.ApiSecret)
	hmac512 := hmac.New(sha512.New, decodedSecret)
	hmac512.Write(append([]byte(urlPath), sha256Sum[:]...))
	return base64.StdEncoding.EncodeToString(hmac512.Sum(nil))
}

func (c *Client) GetBalances() error {
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	data := fmt.Sprintf("nonce=%s", nonce)

	req, err := http.NewRequest("POST", "https://api.kraken.com/0/private/Balance", strings.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Add("API-Key", c.Config.ApiKey)
	req.Header.Add("API-Sign", c.GenerateSignature("/0/private/Balance", data, nonce))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var balanceResp models.BalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return err
	}

	if len(balanceResp.Error) > 0 {
		return fmt.Errorf("API error: %v", balanceResp.Error)
	}

	c.Balances = make(map[string]float64)
	for asset, balStr := range balanceResp.Result {
		if bal, err := utils.ParseFloat(balStr); err == nil && bal > 0 {
			c.Balances[asset] = bal
		}
	}
	return nil
}

func (c *Client) Connect() error {
	if err := c.GetBalances(); err != nil {
		return fmt.Errorf("failed to get balances: %v", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial("wss://ws.kraken.com", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %v", err)
	}
	c.WsConn = conn

	pairs := make([]string, 0)
	for asset := range c.Balances {
		if pair, ok := models.AssetMapping[asset]; ok && pair != "USD" {
			pairs = append(pairs, pair)
		}
	}

	if len(pairs) > 0 {
		msg := map[string]interface{}{
			"event": "subscribe",
			"pair":  pairs,
			"subscription": map[string]interface{}{
				"name": "ticker",
			},
		}
		return c.WsConn.WriteJSON(msg)
	}
	return nil
}

func (c *Client) UpdatePrice(pair string, price float64) {
	c.PrevPrices[pair] = c.Prices[pair]
	c.Prices[pair] = price
}

func (c *Client) GetPrice(pair string) float64 {
	return c.Prices[pair]
}

func (c *Client) GetAssetValues() []models.AssetValue {
	assets := make([]models.AssetValue, 0, len(c.Balances))

	for asset, balance := range c.Balances {
		if asset == "ZUSD" {
			assets = append(assets, models.AssetValue{
				Asset:     "USD",
				Balance:   balance,
				Price:     1.0,
				PrevPrice: 1.0,
				USDValue:  balance,
			})
			continue
		}

		if pair, ok := models.AssetMapping[asset]; ok {
			price := c.Prices[pair]
			prevPrice := c.PrevPrices[pair]
			assets = append(assets, models.AssetValue{
				Asset:     strings.TrimPrefix(strings.TrimPrefix(asset, "X"), "Z"),
				Balance:   balance,
				Price:     price,
				PrevPrice: prevPrice,
				USDValue:  balance * price,
			})
		}
	}

	return assets
}

func (c *Client) StartStreaming(renderFunc func([]models.AssetValue)) {
	for {
		var message json.RawMessage
		if err := c.WsConn.ReadJSON(&message); err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

		var data []interface{}
		if err := json.Unmarshal(message, &data); err == nil && len(data) > 1 {
			if tickerData, ok := data[1].(map[string]interface{}); ok {
				if closeData, ok := tickerData["c"].([]interface{}); ok && len(closeData) > 0 {
					if price, ok := closeData[0].(string); ok {
						if pairInfo, ok := data[3].(string); ok {
							if priceVal, err := utils.ParseFloat(price); err == nil {
								c.UpdatePrice(pairInfo, priceVal)
								assets := c.GetAssetValues()
								renderFunc(assets)
							}
						}
					}
				}
			}
		}
	}
}

func (c *Client) Close() error {
	if c.WsConn != nil {
		return c.WsConn.Close()
	}
	return nil
}
