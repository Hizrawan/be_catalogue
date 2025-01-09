package date

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func IsValidHourFormat(value interface{}) error {
	hourString := value.(string)

	if hourString == "" {
		return nil
	}
	times := strings.Split(hourString, ":")
	if len(times) != 2 {
		return errors.New("time format not supported")
	}

	if len(times[0]) != 2 {
		return errors.New("time format not supported")
	}
	hour, err := strconv.ParseInt(times[0], 10, 64)
	if err != nil {
		return err
	}

	if len(times[1]) != 2 {
		return errors.New("time format not supported")
	}
	minute, err := strconv.ParseInt(times[1], 10, 64)
	if err != nil {
		return err
	}

	if hour < 0 || hour > 23 {
		return errors.New("time format not supported")
	}

	if minute < 0 || minute > 59 {
		return errors.New("time format not supported")
	}

	return nil
}

func IsSubsequentTo(beforeHour string) validation.RuleFunc {
	return func(value interface{}) error {
		afterHour, _ := value.(string)
		if strings.Compare(afterHour, beforeHour) != 1 {
			return fmt.Errorf("%v is not subsequent to %v", afterHour, beforeHour)
		}
		return nil
	}
}
