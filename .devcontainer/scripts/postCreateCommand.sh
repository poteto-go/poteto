#!/bin/bash

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.5
go install github.com/google/pprof@latest
go mod tidy
