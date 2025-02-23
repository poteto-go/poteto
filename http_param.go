package poteto

import (
	"github.com/goccy/go-json"
	"github.com/poteto-go/poteto/constant"
)

type ParamUnit struct {
	key   string
	value string
}

type httpParam struct {
	params map[string]map[string]string `json:"params"`
}

type HttpParam interface {
	GetParam(paramType, key string) (string, bool)
	AddParam(paramType string, paramUnit ParamUnit)
	JsonSerialize() ([]byte, error)
}

func NewHttpParam() HttpParam {
	params := make(map[string]map[string]string, 2)

	httpParam := &httpParam{
		params: params,
	}

	httpParam.params[constant.ParamTypePath] = make(map[string]string)
	httpParam.params[constant.ParamTypeQuery] = make(map[string]string)
	return httpParam
}

func (hp *httpParam) GetParam(paramType, key string) (string, bool) {
	val := hp.params[paramType][key]
	if val != "" {
		return val, true
	}

	return "", false
}

func (hp *httpParam) AddParam(paramType string, paramUnit ParamUnit) {
	hp.params[paramType][paramUnit.key] = paramUnit.value
}

func (hp *httpParam) JsonSerialize() ([]byte, error) {
	v, err := json.Marshal(hp.params)
	if err != nil {
		return []byte{}, err
	}

	return v, nil
}
