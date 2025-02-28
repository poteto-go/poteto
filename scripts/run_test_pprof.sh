#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

LANG=C

go test -cpuprofile cpu.prof -memprofile mem.prof -bench .
