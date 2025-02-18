#!/bin/bash

go install -v github.com/quasilyte/go-ruleguard/cmd/ruleguard@latest
go install github.com/google/pprof@latest
go mod tidy
