#! /bin/bash -x

go fmt ./...
go build -mod=vendor ./...
go vet ./...
