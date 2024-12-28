#!bin/bash
echo "start test"
# noinlineによるテスト実行
# https://github.com/poteto0/poteto/issues/169
go test ./... -coverprofile cover.out.tmp -bench . -benchtime 100000x go test ./... -cover -gcflags=all=-l

echo "remove"
cat cover.out.tmp | grep -v "github.com/poteto0/poteto/cmd/template" > cover2.out.tmp
cat cover2.out.tmp | grep -v "github.com/poteto0/poteto/constant" > coverage.txt

echo "report"
tool cover -func cover.out

echo "post test"
rm -f cover.out.tmp
rm -f cover2.out.tmp
rm -f coverage.txt