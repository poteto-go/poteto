package poteto

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/poteto/utils"
)

type Route interface {
	Search(path string) (*route, []ParamUnit)
	Insert(path string, handler HandlerFunc)

	GetHandler() HandlerFunc
}

type route struct {
	children      map[string]Route
	childParamKey string
	handler       HandlerFunc
}

func NewRoute() Route {
	return &route{
		children:      make(map[string]Route),
		childParamKey: "",
	}
}

func (r *route) Search(path string) (*route, []ParamUnit) {
	currentRoute := r
	rightPath := path[1:]
	param := ""
	httpParams := make([]ParamUnit, 0)

	if rightPath == "" {
		return currentRoute, httpParams
	}

	// optimized router insert
	// https://github.com/poteto-go/poteto/issues/113
	for {
		id := strings.Index(rightPath, "/")
		if id < 0 {
			param = rightPath
		} else {
			param = rightPath[:id]
			rightPath = rightPath[(id + 1):]
		}

		if nextRoute, ok := currentRoute.children[param]; ok {
			currentRoute = nextRoute.(*route)
		} else {
			// includes url param ex: /users/:id, /users/:id/name
			if chParam := currentRoute.childParamKey; chParam != "" {
				if nextRoute, ok = currentRoute.children[chParam]; ok {
					currentRoute = nextRoute.(*route)
					httpParam := ParamUnit{key: chParam, value: param}
					httpParams = append(httpParams, httpParam)
				}
			} else {
				return nil, httpParams
			}
		}

		if id < 0 {
			break
		}
	}

	return currentRoute, httpParams
}

func (r *route) Insert(path string, handler HandlerFunc) {
	currentRoute := r
	rightPath := path[1:]
	param := ""

	// optimized router insert
	// https://github.com/poteto-go/poteto/issues/113
	for {
		id := strings.Index(rightPath, "/")
		if id < 0 { // means last
			param = rightPath
		} else {
			param = rightPath[:id]
			rightPath = rightPath[(id + 1):]
		}

		if nextRoute := currentRoute.children[param]; nextRoute == nil {
			// last path includes url param ex: /users/:id
			if hasParamPrefix(param) {
				currentRoute.childParamKey = param
			}

			currentRoute.children[param] = &route{
				children: make(map[string]Route),
			}
		}
		currentRoute = currentRoute.children[param].(*route)

		if id < 0 {
			break
		}
	}

	if currentRoute.handler != nil {
		coloredWarn := color.HiRedString(fmt.Sprintf("Handler Collision on %s \n", path))
		utils.PotetoPrint(coloredWarn)
		return
	}

	currentRoute.handler = handler
}

func hasParamPrefix(param string) bool {
	return strings.HasPrefix(param, constant.ParamPrefix)
}

func (r *route) GetHandler() HandlerFunc {
	return r.handler
}
