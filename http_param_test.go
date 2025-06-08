package poteto

import (
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/goccy/go-json"
	"github.com/poteto-go/poteto/constant"
	"github.com/stretchr/testify/assert"
)

func TestAddAndGetParam(t *testing.T) {
	hp := NewHttpParam()

	pu := ParamUnit{"key", "value"}
	hp.AddParam(constant.ParamTypePath, pu)

	tests := []struct {
		name         string
		key          string
		expected_val string
		expected_ok  bool
	}{
		{"test ok case", "key", "value", true},
		{"test unexpected", "unexpected", "", false},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			value, ok := hp.GetParam(constant.ParamTypePath, it.key)

			if value != it.expected_val {
				t.Errorf("Don't Work")
			}

			if ok != it.expected_ok {
				t.Errorf("Unmatched")
			}
		})
	}
}

func TestHttpParam_GetParhParam(t *testing.T) {
	hp := NewHttpParam().(*httpParam)
	hp.PathParams["key"] = "value"

	tests := []struct {
		name          string
		expectedVal   string
		expectedFound bool
	}{
		{"key", "value", true},
		{"unexpected", "", false},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			val, found := hp.GetPathParam(it.name)
			assert.Equal(t, it.expectedVal, val)
			assert.Equal(t, it.expectedFound, found)
		})
	}
}

func TestHttpParam_GetQueryParam(t *testing.T) {
	hp := NewHttpParam().(*httpParam)
	hp.QueryParams["key"] = "value"

	tests := []struct {
		name          string
		expectedVal   string
		expectedFound bool
	}{
		{"key", "value", true},
		{"unexpected", "", false},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			val, found := hp.GetQueryParam(it.name)
			assert.Equal(t, it.expectedVal, val)
			assert.Equal(t, it.expectedFound, found)
		})
	}
}

func TestHttpParam_AddPathParam(t *testing.T) {
	hp := NewHttpParam().(*httpParam)

	tests := []struct {
		name string
		key  string
		val  string
	}{
		{"test", "key", "value"},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			hp.AddPathParam(ParamUnit{key: it.key, value: it.val})
			val, _ := hp.PathParams[it.key]
			assert.Equal(t, it.val, val)
		})
	}
}

func TestHttpParam_AddQueryParam(t *testing.T) {
	hp := NewHttpParam().(*httpParam)

	tests := []struct {
		name string
		key  string
		val  string
	}{
		{"test", "key", "value"},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			hp.AddQueryParam(ParamUnit{key: it.key, value: it.val})
			val, _ := hp.QueryParams[it.key]
			assert.Equal(t, it.val, val)
		})
	}
}

func TestJsonSerializeHttpParam(t *testing.T) {
	hp := NewHttpParam()
	hp.AddParam(constant.ParamTypePath, ParamUnit{key: "key", value: "value"})

	expected := `{"path":{"key":"value"},"query":{}}`
	serialized, _ := hp.JsonSerialize()
	if string(serialized) != expected {
		t.Errorf(
			"Unmatched actual(%s) -> expected(%s)",
			string(serialized),
			expected,
		)
	}
}

func TestJsonSerializeHttpHandleError(t *testing.T) {
	defer monkey.UnpatchAll()

	hp := NewHttpParam()
	monkey.Patch(json.Marshal, func(v any) ([]byte, error) {
		return []byte(""), errors.New("error")
	})

	if _, err := hp.JsonSerialize(); err == nil {
		t.Errorf("Unmatched")
	}
}
