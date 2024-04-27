@echo off

set GOARCH=amd64
set GOOS=linux

garble.exe -literals build -o ./release/Tempo .

set GOARCH=
set GOOS=