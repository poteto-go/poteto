package utils

import (
	"errors"
	"testing"
)

type AB struct {
	A string `yaml:"a"`
	B string `yaml:"b"`
}

var data = `
a: test
b: hello
`

// tab not allowed
var not_expected = `
	a: *
	b: "hello"
`

func TestYamlParse(t *testing.T) {
	tests := []struct {
		name      string
		yaml_file any
		worked    bool
		expected  AB
	}{
		{"Test string yaml", data, true, AB{A: "test", B: "hello"}},
		{"Test []byte yaml", []byte(data), true, AB{A: "test", B: "hello"}},
		{"Test not string | []byte yaml", 1, false, AB{A: "test", B: "hello"}},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			var ab AB

			switch asserted := any(it.yaml_file).(type) {
			case string:
				if err := YamlParse(asserted, &ab); err != nil {
					if it.worked {
						t.Errorf("Not expected Error")
					}
					return
				}

				if !it.worked {
					t.Errorf("Not occurred error")
					return
				}

				if it.expected.A != ab.A || it.expected.B != ab.B {
					t.Errorf("Not matched")
				}

			case []byte:
				if err := YamlParse(asserted, &ab); err != nil {
					if it.worked {
						t.Errorf("Not expected Error")
					}
					return
				}

				if !it.worked {
					t.Errorf("Not occurred error")
					return
				}

				if it.expected.A != ab.A || it.expected.B != ab.B {
					t.Errorf("Not matched")
				}
			}
		})
	}
}

func TestYamlParseNotPtrCase(t *testing.T) {
	var ab AB
	yamlFile := []byte(data)
	err := YamlParse(yamlFile, ab)
	if err == nil {
		t.Errorf("Not throw error")
	}
}

func TestAssertToInt(t *testing.T) {
	tests := []struct {
		name     string
		source   any
		expected bool
	}{
		{
			"Test int case",
			int(10),
			true,
		},
		{
			"Test string case",
			string("10"),
			true,
		},
		{
			"Test byteArray case",
			[]byte("10"),
			true,
		},
		{
			"Test float64 case",
			float64(10),
			true,
		},
		{
			"Test invalid case",
			errors.New("error"),
			false,
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			_, ok := AssertToInt(it.source)
			if ok != it.expected {
				t.Errorf("Unmatched actual(%v) -> expected(%v)", ok, it.expected)
			}
		})
	}
}
