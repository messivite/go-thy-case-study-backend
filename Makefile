.PHONY: openapi-json openapi-validate build test

openapi-json: docs/openapi.yaml
	@echo "generating docs/openapi.json ..."
	@go run ./scripts/openapi_json.go

openapi-validate: openapi-json
	@echo "validating openapi spec drift ..."
	@git diff --exit-code docs/openapi.json || (echo "ERROR: docs/openapi.json out of date — run 'make openapi-json' and commit" && exit 1)

build:
	go build ./...

test:
	go test ./... -v -count=1
