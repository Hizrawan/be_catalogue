package date

import (
	"fmt"
	"time"
)

func RocYear(date time.Time) int {
	return date.Year() - 1911
}

func RocDate(date time.Time) string {
	year := RocYear(date)
	return fmt.Sprintf("%d%s", year, date.Format("0102"))
}

func RocDate7(date time.Time) string {
	return fmt.Sprintf("%07s", RocDate(date))
}
