package utils

const (
	USD = "USD"
	EUR = "EUR"
	GBP = "GBP"
	NGN = "NGN"
	AUD = "AUD"
	CAD = "CAD"
	CDF = "CDF"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD, CDF, AUD, NGN, GBP:
		return true
	}
	return false
}
