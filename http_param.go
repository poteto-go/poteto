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
	pathParams  map[string]string
	queryParams map[string]string
}

type HttpParam interface {
	selectParam(paramType string) map[string]string
	GetParam(paramType, key string) (string, bool)
	AddParam(paramType string, paramUnit ParamUnit)
	JsonSerialize() ([]byte, error)
}

func NewHttpParam() HttpParam {
	httpParam := &httpParam{
		pathParams:  make(map[string]string),
		queryParams: make(map[string]string),
	}

	return httpParam
}

func (hp *httpParam) GetParam(paramType, key string) (string, bool) {
	targetParams := hp.selectParam(paramType)
	val := targetParams[key]
	if val != "" {
		return val, true
	}

	return "", false
}

func (hp *httpParam) AddParam(paramType string, paramUnit ParamUnit) {
	targetParams := hp.selectParam(paramType)
	targetParams[paramUnit.key] = paramUnit.value
}

func (hp *httpParam) selectParam(paramType string) map[string]string {
	switch paramType {
	case constant.ParamTypePath:
		return hp.pathParams
	case constant.ParamTypeQuery:
		return hp.queryParams
	}
	return make(map[string]string)
}

func (hp *httpParam) JsonSerialize() ([]byte, error) {
	unionParams := map[string]map[string]string{
		constant.ParamTypePath:  hp.pathParams,
		constant.ParamTypeQuery: hp.queryParams,
	}

	v, err := json.Marshal(unionParams)
	if err != nil {
		return []byte{}, err
	}

	return v, nil
}
