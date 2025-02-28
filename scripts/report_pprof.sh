#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

LANG=C

go tool pprof -top cpu.prof | grep github.com/poteto-go
go tool pprof -top mem.prof | grep github.com/poteto-go
