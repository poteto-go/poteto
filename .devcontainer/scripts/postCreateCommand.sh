#!/bin/bash

go install -v github.com/quasilyte/go-ruleguard/cmd/ruleguard@latest
go mod tidy
