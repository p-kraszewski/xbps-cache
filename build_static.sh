#!/bin/sh

CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w'
