package utils

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/goccy/go-yaml"
)

func YamlParse[T string | []byte](source T, dest any) error {
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return errors.New("dest is not pointer")
	}

	switch source_asserted := any(source).(type) {
	case string:
		parseError := yaml.Unmarshal([]byte(source_asserted), dest)
		return parseError
	case []byte:
		parseError := yaml.Unmarshal(source_asserted, dest)
		return parseError
	default:
		return errors.New("unexpected input, expected string | []byte")
	}
}

func AssertToInt(source any) (int, bool) {
	switch asserted := source.(type) {
	case string:
		tmp, _ := strconv.Atoi(asserted)
		return tmp, true
	case int:
		return asserted, true
	case []byte:
		tmp, _ := strconv.Atoi(string(asserted))
		return tmp, true
	case float64:
		return int(asserted), true
	default:
		return 0, false
	}
}
