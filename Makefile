.PHONY: build

fmt:
	@go fmt ./...

tidy:
	@go mod tidy

build: tidy
	@go build -o packer-plugin-testinfra

install: build
	@packer plugins install --path ./packer-plugin-testinfra 'github.com/mschuchard/testinfra'

unit:
	@go test -v ./provisioner

accept: install
	# start vbox machine for ssh communicator testing
	@PACKER_ACC=1 go test -v ./main_test.go -timeout=1m

install-packer-sdc:
	@go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@latest

#release-docs: install-packer-sdc
#	@packer-sdc renderdocs -src docs -partials docs-partials/ -dst docs/
#	@/bin/sh -c "[ -d docs ] && zip -r docs.zip docs/"

plugin-check: install-packer-sdc build
	@~/go/bin/packer-sdc plugin-check packer-plugin-testinfra

generate: install-packer-sdc
	export PATH="${PATH}:$(shell go env GOPATH)/bin"
	@go generate ./...
	#packer-sdc renderdocs -src ./docs -dst ./.docs -partials ./docs-partials

test: unit accept plugin-check
