package stringsutil

import (
	"strconv"
	"strings"
)

func QuoteStrBySep(str string, sep string) string {
	queries := strings.Split(str, sep)

	for i := range queries {
		queries[i] = strconv.Quote(queries[i])
	}

	return strings.Join(queries, sep)
}

func NormalizeLangForTemplate(lang string) string {
	return strings.ReplaceAll(strings.ToLower(lang), "-", "_")
}
