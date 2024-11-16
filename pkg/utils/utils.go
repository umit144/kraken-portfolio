package utils

import (
	"fmt"
)

func ParseFloat(s string) (float64, error) {
	var val float64
	_, err := fmt.Sscanf(s, "%f", &val)
	return val, err
}

func FormatFloat(f float64, precision int) string {
	return fmt.Sprintf("%.*f", precision, f)
}
