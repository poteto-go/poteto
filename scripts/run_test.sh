#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

LANG=C

echo "start test"
# noinlineによるテスト実行
# https://github.com/poteto-go/poteto/issues/169
go test ./... -cover -coverprofile cover.out.tmp -bench . -benchtime 100000x -gcflags=all=-l

echo "remove"
cat cover.out.tmp | grep -v "github.com/poteto-go/poteto/cmd/template" > cover2.out.tmp
cat cover2.out.tmp | grep -v "github.com/poteto-go/poteto/constant" > coverage.txt

echo "report"
go tool cover -func cover.out

echo "post test"
rm -f cover.out.tmp
rm -f cover2.out.tmp
rm -f coverage.txt
