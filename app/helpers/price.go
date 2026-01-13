package helpers

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func FormatPrice(price interface{}) string {
	switch v := price.(type) {
	case decimal.Decimal:
		return fmt.Sprintf("Rp %s", v.StringFixed(0))
	case *decimal.Decimal:
		if v != nil {
			return fmt.Sprintf("Rp %s", v.StringFixed(0))
		}
		return "Rp 0"
	case int, int64, float32, float64:
		return fmt.Sprintf("Rp %v", v)
	default:
		return "Rp 0"
	}
}
