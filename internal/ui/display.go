package ui

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/umit144/kraken-portfolio/internal/models"

	"golang.org/x/term"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorCyan  = "\033[36m"
	colorGray  = "\033[37m"
	bgBlack    = "\033[40m"

	minWidth = 60
	maxWidth = 100

	assetWidth   = 6
	balanceWidth = 10
	priceWidth   = 12
	valueWidth   = 12
)

type Display struct {
	width  int
	writer io.Writer
}

func calculateWidth(requestedWidth int) int {
	if requestedWidth < minWidth {
		return minWidth
	}
	if requestedWidth > maxWidth {
		return maxWidth
	}
	return requestedWidth
}

func NewDisplay() *Display {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
	}
	return &Display{
		width:  calculateWidth(width),
		writer: os.Stdout,
	}
}

func NewDisplayWithWriter(w io.Writer, width int) *Display {
	return &Display{
		width:  calculateWidth(width),
		writer: w,
	}
}

func (d *Display) GetPriceColor(current, previous float64) string {
	if current > previous {
		return colorGreen
	} else if current < previous {
		return colorRed
	}
	return colorReset
}

func (d *Display) FormatPrice(price float64, color string) string {
	return fmt.Sprintf("%s$%.2f%s", color, price, colorReset)
}

func (d *Display) FormatBalance(balance float64) string {
	if balance >= 1000 {
		return fmt.Sprintf("%.2f", balance)
	}
	return fmt.Sprintf("%.8f", balance)
}

func (d *Display) RenderPortfolio(assets []models.AssetValue) {
	fmt.Fprint(d.writer, "\033[H\033[2J")
	d.renderHeader()

	cryptoAssets, usdAsset := d.separateAssets(assets)
	d.renderCryptoAssets(cryptoAssets)

	if usdAsset != nil {
		d.renderDivider()
		d.renderUSD(*usdAsset)
	}

	totalUSD := d.calculateTotal(assets)
	d.renderFooter(totalUSD)
}

func (d *Display) separateAssets(assets []models.AssetValue) ([]models.AssetValue, *models.AssetValue) {
	var cryptoAssets []models.AssetValue
	var usdAsset *models.AssetValue

	for _, asset := range assets {
		if asset.Asset == "USD" {
			assetCopy := asset
			usdAsset = &assetCopy
		} else {
			cryptoAssets = append(cryptoAssets, asset)
		}
	}

	sort.Slice(cryptoAssets, func(i, j int) bool {
		return cryptoAssets[i].USDValue > cryptoAssets[j].USDValue
	})

	return cryptoAssets, usdAsset
}

func (d *Display) calculateTotal(assets []models.AssetValue) float64 {
	total := 0.0
	for _, asset := range assets {
		total += asset.USDValue
	}
	return total
}

func (d *Display) renderHeader() {
	title := "KRAKEN PORTFOLIO"
	titlePadding := (d.width - len(title)) / 2

	fmt.Fprintf(d.writer, "%s%s╔%s╗%s\n",
		bgBlack, colorCyan, strings.Repeat("═", d.width), colorReset)

	fmt.Fprintf(d.writer, "%s║%s%s%s║%s\n",
		colorCyan,
		strings.Repeat(" ", titlePadding),
		title,
		strings.Repeat(" ", d.width-len(title)-titlePadding),
		colorReset,
	)

	fmt.Fprintf(d.writer, "%s╠%s╣%s\n",
		colorCyan, strings.Repeat("═", d.width), colorReset)

	fmt.Fprintf(d.writer, "%s║ %-*s %-*s %-*s %-*s ║%s\n",
		colorCyan,
		assetWidth, "ASSET",
		balanceWidth, "BALANCE",
		priceWidth, "PRICE",
		valueWidth, "VALUE (USD)",
		colorReset)

	fmt.Fprintf(d.writer, "%s╠%s╣%s\n",
		colorCyan, strings.Repeat("═", d.width), colorReset)
}

func (d *Display) renderCryptoAssets(assets []models.AssetValue) {
	for _, asset := range assets {
		priceColor := d.GetPriceColor(asset.Price, asset.PrevPrice)
		priceStr := d.FormatPrice(asset.Price, priceColor)
		balanceStr := d.FormatBalance(asset.Balance)

		fmt.Fprintf(d.writer, "%s║ %-*s %-*s %-*s %-*.2f ║%s\n",
			colorCyan,
			assetWidth, asset.Asset,
			balanceWidth, balanceStr,
			priceWidth, priceStr,
			valueWidth, asset.USDValue,
			colorReset)
	}
}

func (d *Display) renderDivider() {
	fmt.Fprintf(d.writer, "%s╟%s╢%s\n",
		colorCyan, strings.Repeat("─", d.width), colorReset)
}

func (d *Display) renderUSD(usd models.AssetValue) {
	fmt.Fprintf(d.writer, "%s║ %-*s %-*.2f %-*s %-*.2f ║%s\n",
		colorCyan,
		assetWidth, usd.Asset,
		balanceWidth, usd.Balance,
		priceWidth, "-",
		valueWidth, usd.USDValue,
		colorReset)
}

func (d *Display) renderFooter(totalUSD float64) {
	fmt.Fprintf(d.writer, "%s╠%s╣%s\n",
		colorCyan, strings.Repeat("═", d.width), colorReset)

	fmt.Fprintf(d.writer, "%s║ TOTAL VALUE: $%-*.*f ║%s\n",
		colorCyan, d.width-15, 2, totalUSD, colorReset)

	fmt.Fprintf(d.writer, "%s╚%s╝%s\n",
		colorCyan, strings.Repeat("═", d.width), colorReset)

	fmt.Fprintf(d.writer, "%sPress Ctrl+C to exit%s\n",
		colorGray, colorReset)
}
