package util

const (
	EUR = "EUR"
	USD = "USD"
	BRL = "BRL"
)

// IsSupportedCurrency returns true if a currrency is supported
func IsSupportedCurrency(c string) bool {
	switch c {
	case EUR, USD, BRL:
		return true
	}
	return false
}
