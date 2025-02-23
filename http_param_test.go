package poteto

import (
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/goccy/go-json"
	"github.com/poteto-go/poteto/constant"
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
