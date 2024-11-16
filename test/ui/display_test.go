package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/umit144/kraken-portfolio/internal/models"
	"github.com/umit144/kraken-portfolio/internal/ui"

	"github.com/stretchr/testify/assert"
)

func TestNewDisplay(t *testing.T) {
	display := ui.NewDisplay()
	assert.NotNil(t, display)
}

func TestGetPriceColor(t *testing.T) {
	display := ui.NewDisplayWithWriter(nil, 80)
	tests := []struct {
		name     string
		current  float64
		previous float64
		want     string
	}{
		{
			name:     "price increase",
			current:  3000.0,
			previous: 2900.0,
			want:     "\033[32m",
		},
		{
			name:     "price decrease",
			current:  2800.0,
			previous: 2900.0,
			want:     "\033[31m",
		},
		{
			name:     "price unchanged",
			current:  3000.0,
			previous: 3000.0,
			want:     "\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := display.GetPriceColor(tt.current, tt.previous)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatPrice(t *testing.T) {
	display := ui.NewDisplayWithWriter(nil, 80)
	tests := []struct {
		name  string
		price float64
		color string
		want  string
	}{
		{
			name:  "positive price green",
			price: 3000.0,
			color: "\033[32m",
			want:  "\033[32m$3000.00\033[0m",
		},
		{
			name:  "positive price red",
			price: 2900.0,
			color: "\033[31m",
			want:  "\033[31m$2900.00\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := display.FormatPrice(tt.price, tt.color)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatBalance(t *testing.T) {
	display := ui.NewDisplayWithWriter(nil, 80)
	tests := []struct {
		name    string
		balance float64
		want    string
	}{
		{
			name:    "large balance",
			balance: 1234.5678,
			want:    "1234.57",
		},
		{
			name:    "small balance",
			balance: 0.12345678,
			want:    "0.12345678",
		},
		{
			name:    "zero balance",
			balance: 0.0,
			want:    "0.00000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := display.FormatBalance(tt.balance)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRenderPortfolio(t *testing.T) {
	var buf bytes.Buffer
	display := ui.NewDisplayWithWriter(&buf, 80)

	assets := []models.AssetValue{
		{
			Asset:     "ETH",
			Balance:   1.5,
			Price:     3000.0,
			PrevPrice: 2900.0,
			USDValue:  4500.0,
		},
		{
			Asset:     "SOL",
			Balance:   10.0,
			Price:     100.0,
			PrevPrice: 110.0,
			USDValue:  1000.0,
		},
		{
			Asset:     "USD",
			Balance:   1000.0,
			Price:     1.0,
			PrevPrice: 1.0,
			USDValue:  1000.0,
		},
	}

	display.RenderPortfolio(assets)
	output := buf.String()

	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "contains headers",
			check: func(t *testing.T) {
				assert.Contains(t, output, "KRAKEN PORTFOLIO")
				assert.Contains(t, output, "ASSET")
				assert.Contains(t, output, "BALANCE")
				assert.Contains(t, output, "PRICE")
				assert.Contains(t, output, "VALUE (USD)")
			},
		},
		{
			name: "contains assets",
			check: func(t *testing.T) {
				assert.Contains(t, output, "ETH")
				assert.Contains(t, output, "SOL")
				assert.Contains(t, output, "USD")
			},
		},
		{
			name: "contains formatted values",
			check: func(t *testing.T) {
				assert.Contains(t, output, "1.50000000")
				assert.Contains(t, output, "$3000.00")
				assert.Contains(t, output, "4500.00")
			},
		},
		{
			name: "correct order by value",
			check: func(t *testing.T) {
				ethIndex := strings.Index(output, "ETH")
				solIndex := strings.Index(output, "SOL")
				assert.True(t, ethIndex < solIndex, "ETH should appear before SOL")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.check)
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		maxLength int
	}{
		{"minimum width", 40, 60},
		{"normal width", 80, 80},
		{"maximum width", 120, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			display := ui.NewDisplayWithWriter(&buf, tt.width)

			assets := []models.AssetValue{
				{
					Asset:    "ETH",
					Balance:  1.0,
					Price:    3000.0,
					USDValue: 3000.0,
				},
			}

			display.RenderPortfolio(assets)
			output := buf.String()

			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if len(line) == 0 {
					continue
				}

				cleanLine := removeANSICodes(line)
				assert.LessOrEqual(t,
					len(cleanLine),
					tt.maxLength,
					"line length (without color codes) should not exceed maximum width: %s", line,
				)
			}
		})
	}
}

func removeANSICodes(s string) string {
	return strings.Join(strings.Split(s, "\033[")[0:1], "")
}
