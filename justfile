#!/usr/bin/env just --justfile

generate:
	go generate ./...
	oapi-codegen --config oapi.cfg.yaml openapi.yaml

dockerfile:
	goctl docker -go main.go --tz Europe/Moscow
