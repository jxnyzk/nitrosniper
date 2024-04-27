@echo off

set GOARCH=amd64
set GOOS=linux

go build -o dev/Tempo .

set GOARCH=
set GOOS=