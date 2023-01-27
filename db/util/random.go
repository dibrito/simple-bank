package util

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var alphabet = "abcdefghijklmnopqrstuvxzyw"

// RandomInt generates a random integer between min and maxx
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a randpm string of lenght n
func RandomString(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(len(alphabet))]
		sb.WriteByte(c)
	}
	return sb.String()
}

// RandomOwner generates a random owner name
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money
func RandonMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomMoney generates a random amount of money
func RandonCurrency() string {
	c := []string{"USD", "EUR", "BRL"}
	return c[rand.Intn(len(c))]
}
