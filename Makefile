.PHONY: openapi openapi-validate build test

openapi: api.yaml
	@echo "generating docs/openapi.yaml and docs/openapi.json from api.yaml ..."
	@go run ./scripts/generate_openapi.go

openapi-validate: openapi
	@echo "validating openapi spec drift ..."
	@git diff --exit-code docs/openapi.yaml docs/openapi.json || (echo "ERROR: OpenAPI docs out of date — run 'make openapi' and commit" && exit 1)

build:
	go build ./...

test:
	go test ./... -v -count=1
