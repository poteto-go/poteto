package utils

import (
	"path"
	"strings"

	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/poteto/perror"
)

func buildUrl(basePath, addPath string) string {
	if !strings.Contains(basePath, "/") {
		basePath = "/" + basePath
	}

	if !strings.Contains(addPath, "/") {
		addPath = "/" + addPath
	}

	fullPath := path.Join(basePath, addPath)

	if fullPath == "" {
		return "/"
	}

	return path.Clean(fullPath)
}

func BuildSafeUrl(basePath, addPath string) (string, error) {
	if strings.Contains(basePath, "..") {
		return "", perror.ErrPathTraversalNotAllowed
	}

	if strings.Contains(addPath, "..") {
		return "", perror.ErrPathTraversalNotAllowed
	}

	fullPath := buildUrl(basePath, addPath)
	if len(fullPath) > constant.MAX_DOMAIN_LENGTH {
		return "", perror.ErrPathLengthExceeded
	}

	return fullPath, nil
}
