package algorithm

type HandlerGraph interface {
	Children() map[string]HandlerGraph
	Handler() any
}

type SearchResult struct {
	path    string
	handler any
}

func DepthFirstSearch(root HandlerGraph) []SearchResult {
	results := []SearchResult{}

	visited := map[string]struct{}{}
	depthFirstSearch(root, "", &visited, &results)

	return results
}

func depthFirstSearch(
	root HandlerGraph,
	path string,
	visited *map[string]struct{},
	results *[]SearchResult,
) {
	if _, ok := (*visited)[path]; ok {
		return
	}

	(*visited)[path] = struct{}{}
	for key, child := range root.Children() {
		if child.Handler() != nil {
			*results = append(
				*results,
				SearchResult{path, child.Handler()},
			)
		}

		nextPath := key + path
		depthFirstSearch(child, nextPath, visited, results)
	}
}
