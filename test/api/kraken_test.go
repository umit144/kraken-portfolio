package api_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/umit144/kraken-portfolio/internal/api"
	"github.com/umit144/kraken-portfolio/internal/config"
	"github.com/umit144/kraken-portfolio/internal/models"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		ApiKey:    "test-key",
		ApiSecret: "test-secret",
	}

	client := api.NewClient(cfg)
	if client == nil {
		t.Error("Expected non-nil client")
	}
	if client.Prices == nil {
		t.Error("Expected non-nil prices map")
	}
	if client.PrevPrices == nil {
		t.Error("Expected non-nil previous prices map")
	}
	if client.Balances == nil {
		t.Error("Expected non-nil balances map")
	}
}

func TestGenerateSignature(t *testing.T) {
	cfg := &config.Config{
		ApiKey:    "test-key",
		ApiSecret: base64.StdEncoding.EncodeToString([]byte("test-secret")),
	}

	client := api.NewClient(cfg)
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	signature := client.GenerateSignature("/0/private/test", "nonce="+nonce, nonce)

	if signature == "" {
		t.Error("Expected non-empty signature")
	}
}

func TestUpdatePrice(t *testing.T) {
	cfg := &config.Config{
		ApiKey:    "test-key",
		ApiSecret: "test-secret",
	}

	client := api.NewClient(cfg)
	pair := "ETH/USD"

	tests := []struct {
		name     string
		price    float64
		expected float64
	}{
		{"initial price", 3000.0, 3000.0},
		{"price increase", 3100.0, 3100.0},
		{"price decrease", 2900.0, 2900.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.UpdatePrice(pair, tt.price)
			got := client.GetPrice(pair)
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPricePrevTracking(t *testing.T) {
	cfg := &config.Config{
		ApiKey:    "test-key",
		ApiSecret: "test-secret",
	}

	client := api.NewClient(cfg)
	pair := "ETH/USD"

	initialPrice := 3000.0
	client.UpdatePrice(pair, initialPrice)
	if got := client.GetPrice(pair); got != initialPrice {
		t.Errorf("got %v, want %v", got, initialPrice)
	}

	newPrice := 3100.0
	client.UpdatePrice(pair, newPrice)
	if got := client.GetPrice(pair); got != newPrice {
		t.Errorf("got %v, want %v", got, newPrice)
	}
	if got := client.PrevPrices[pair]; got != initialPrice {
		t.Errorf("got %v, want %v", got, initialPrice)
	}
}

func TestGetAssetValues(t *testing.T) {
	cfg := &config.Config{
		ApiKey:    "test-key",
		ApiSecret: "test-secret",
	}

	client := api.NewClient(cfg)

	client.Balances = map[string]float64{
		"XETH": 2.5,
		"SOL":  10.0,
		"ZUSD": 1000.0,
	}

	client.UpdatePrice("ETH/USD", 3000.0)
	client.UpdatePrice("SOL/USD", 100.0)

	assets := client.GetAssetValues()
	if len(assets) == 0 {
		t.Error("Expected non-empty assets slice")
	}

	assetMap := make(map[string]models.AssetValue)
	for _, asset := range assets {
		assetMap[asset.Asset] = asset
	}

	eth := assetMap["ETH"]
	if eth.Balance != 2.5 || eth.Price != 3000.0 || eth.USDValue != 7500.0 {
		t.Errorf("Unexpected ETH values: %+v", eth)
	}

	sol := assetMap["SOL"]
	if sol.Balance != 10.0 || sol.Price != 100.0 || sol.USDValue != 1000.0 {
		t.Errorf("Unexpected SOL values: %+v", sol)
	}

	usd := assetMap["USD"]
	if usd.Balance != 1000.0 || usd.Price != 1.0 || usd.USDValue != 1000.0 {
		t.Errorf("Unexpected USD values: %+v", usd)
	}
}

func TestConnect(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test")
	}

	cfg := &config.Config{
		ApiKey:    os.Getenv("KRAKEN_API_KEY"),
		ApiSecret: os.Getenv("KRAKEN_API_SECRET"),
	}

	client := api.NewClient(cfg)
	err := client.Connect()
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}
	defer client.Close()

	time.Sleep(2 * time.Second)
	if client.WsConn == nil {
		t.Error("Expected non-nil WebSocket connection")
	}
}

func TestStreamingUpdates(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test")
	}

	cfg := &config.Config{
		ApiKey:    os.Getenv("KRAKEN_API_KEY"),
		ApiSecret: os.Getenv("KRAKEN_API_SECRET"),
	}

	client := api.NewClient(cfg)
	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	updates := make(chan []models.AssetValue, 1)
	go func() {
		client.StartStreaming(func(assets []models.AssetValue) {
			select {
			case updates <- assets:
			default:
			}
		})
	}()

	select {
	case assets := <-updates:
		if len(assets) == 0 {
			t.Error("Expected non-empty assets update")
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for price updates")
	}
}

func TestAssetMappings(t *testing.T) {
	expectedPairs := map[string]string{
		"XETH": "ETH/USD",
		"SOL":  "SOL/USD",
		"XXBT": "XBT/USD",
	}

	for asset, expectedPair := range expectedPairs {
		if pair, ok := models.AssetMapping[asset]; !ok {
			t.Errorf("Missing mapping for asset: %s", asset)
		} else if pair != expectedPair {
			t.Errorf("Wrong pair mapping for %s: expected %s, got %s",
				asset, expectedPair, pair)
		}
	}
}

func TestUSDHandling(t *testing.T) {
	cfg := &config.Config{
		ApiKey:    "test-key",
		ApiSecret: "test-secret",
	}

	client := api.NewClient(cfg)
	client.Balances["ZUSD"] = 1000.0

	assets := client.GetAssetValues()
	var foundUSD bool
	for _, asset := range assets {
		if asset.Asset == "USD" {
			foundUSD = true
			if asset.Balance != 1000.0 || asset.Price != 1.0 || asset.USDValue != 1000.0 {
				t.Errorf("Unexpected USD values: %+v", asset)
			}
			break
		}
	}

	if !foundUSD {
		t.Error("USD asset not found in results")
	}
}
