#!bin/bash
go tool pprof -top cpu.prof | grep github.com/poteto-go
