SHELL := bash

test:
	go test -v ./... -test.short

install:
	go build -o kromium .
