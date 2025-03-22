package poteto

import (
	"strings"
)

/*
 * Middleware Group
 * Trie Tree of path -> middleware
 * If you want to apply into all path, Apply to "" path
 * Refer, if once found.
 */

type middlewareLinear struct {
	path    string
	handler MiddlewareFunc
}

type MiddlewareTree interface {
	SearchMiddlewares(pattern string) []MiddlewareFunc
	Insert(pattern string, middlewares ...MiddlewareFunc) *middlewareTree
	Register(middlewares ...MiddlewareFunc)

	// DFS route & return linearRouter
	//
	// []{
	//   path: string,
	//   handler: MiddlewareFunc,
	// }
	DFS() []middlewareLinear
	dfs(node *middlewareTree, path string, visited *map[string]struct{}, results *[]middlewareLinear)
}

type middlewareTree struct {
	children    map[string]MiddlewareTree
	middlewares []MiddlewareFunc
	key         string
}

func NewMiddlewareTree() MiddlewareTree {
	return &middlewareTree{
		children: make(map[string]MiddlewareTree),
	}
}

func (mt *middlewareTree) SearchMiddlewares(pattern string) []MiddlewareFunc {
	currentNode := mt
	// faster
	middlewares := mt.middlewares
	if pattern == "/" {
		return middlewares
	}

	rightPattern := pattern[1:]
	param := ""

	for {
		id := strings.Index(rightPattern, "/")
		if id < 0 {
			param = rightPattern
		} else {
			param = rightPattern[:id]
			rightPattern = rightPattern[(id + 1):]
		}

		if nextNode, ok := currentNode.children[param]; ok {
			currentNode = nextNode.(*middlewareTree)
			middlewares = append(middlewares, currentNode.middlewares...)
		} else {
			// if found ever
			// You got Middleware Tree
			break
		}
	}
	return middlewares
}

func (mt *middlewareTree) Insert(pattern string, middlewares ...MiddlewareFunc) *middlewareTree {
	currentNode := mt
	if pattern == "/" || pattern == "" {
		currentNode.Register(middlewares...)
		return currentNode
	}
	rightPattern := pattern[1:]
	param := ""

	for {
		id := strings.Index(rightPattern, "/")
		if id < 0 {
			param = rightPattern
		} else {
			param = rightPattern[:id]
			rightPattern = rightPattern[(id + 1):]
		}

		if _, ok := currentNode.children[param]; !ok {
			currentNode.children[param] = &middlewareTree{
				children:    make(map[string]MiddlewareTree),
				middlewares: []MiddlewareFunc{},
				key:         param,
			}
		}
		currentNode = currentNode.children[param].(*middlewareTree)

		if id < 0 {
			break
		}
	}
	currentNode.Register(middlewares...)
	return currentNode
}

func (mt *middlewareTree) Register(middlewares ...MiddlewareFunc) {
	mt.middlewares = append(mt.middlewares, middlewares...)
}

func (mt *middlewareTree) DFS() []middlewareLinear {
	results := make([]middlewareLinear, 0)
	visited := map[string]struct{}{}
	mt.dfs(mt, "", &visited, &results)
	return results
}

func (mt *middlewareTree) dfs(
	node *middlewareTree,
	path string,
	visited *map[string]struct{},
	results *[]middlewareLinear,
) {
	if node == nil {
		return
	}

	if _, ok := (*visited)[path]; ok {
		return
	}

	(*visited)[path] = struct{}{}

	for _, middleware := range node.middlewares {
		if middleware == nil {
			continue
		}

		*results = append(*results, middlewareLinear{
			path:    path,
			handler: middleware,
		})
	}

	for key, child := range node.children {
		nextPath := path + "/" + key
		mt.dfs(
			child.(*middlewareTree),
			nextPath,
			visited,
			results,
		)
	}
}
