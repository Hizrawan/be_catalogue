package mapstructure

import (
	"reflect"
	"time"

	"gopkg.in/guregu/null.v4"
)

func HookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String && f.Kind() != reflect.Int && f.Kind() != reflect.Float64 {
		return data, nil
	}

	var s null.String
	if t == reflect.TypeOf(s) {
		s = null.NewString(data.(string), true)
		return s, nil
	}

	var i null.Int
	if t == reflect.TypeOf(i) {
		i = null.NewInt(int64(data.(float64)), true)
		return i, nil
	}

	var tm null.Time
	if t == reflect.TypeOf(tm) {
		dataStr, _ := data.(string)
		dataTime, _ := time.Parse("2006-01-02T15:04:05Z", dataStr)
		tm = null.NewTime(dataTime, true)
		return tm, nil
	}

	return data, nil
}
