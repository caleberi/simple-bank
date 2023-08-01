package utils

import (
	"math/rand"
	"strings"
	"time"
)

var random *rand.Rand

const alphabets = "ABCDEFGHIJKLMNOPQRSTUVWXYabcdefghijklmnopqrstuvwxyz"

func init() {
	randSrc := rand.NewSource(time.Now().Unix())
	random = rand.New(randSrc)
}

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + random.Int63n(max-min+1)
}

// RandomString generate a random string of length n characters
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabets)
	for i := 0; i < n; i++ {
		c := alphabets[random.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

// RandomOwner returns a random owner name
func RandomOwner() string {
	return RandomString(8)
}

// RandomMoney returns a random money amount
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrencyCode returns a random currency code
func RandomCurrencyCode() string {
	currencyCodes := []string{
		"USD", "EUR",
		"GBP", "NGN",
		"AUD", "CAD", "CDF",
	}
	return currencyCodes[random.Intn(len(currencyCodes))]
}
