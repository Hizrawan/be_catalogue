package random

import (
	"math/rand"
	"strings"
	"time"
)

const NumericCharset = "0123456789"
const UppercaseAlphabeticCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const LowercaseAlphabeticCharset = "abcdefghijklmnopqrstuvwxyz"

var rnd = rand.New(rand.NewSource(time.Now().Unix()))

func GenerateString(length int, charset string) string {
	var str strings.Builder
	set := []rune(charset)
	for i := 0; i < length; i++ {
		idx := rnd.Intn(len(set))
		str.WriteRune(set[idx])
	}
	return str.String()
}
