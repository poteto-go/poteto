#!bin/bash

go test -coverprofile cover.out.tmp -cpuprofile cpu.prof -bench .
