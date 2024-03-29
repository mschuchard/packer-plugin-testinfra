.PHONY: build

fmt:
	@go fmt ./...

tidy:
	@go mod tidy

build: tidy
	@go build -o packer-plugin-testinfra

unit:
	@go test -v ./...

accept: build
	# vagrant up and vagrant suspend one-time in fixtures dir
	@mkdir -p ~/.packer.d/plugins/
	# https://github.com/hashicorp/packer/issues/11972
	@cp packer-plugin-testinfra ~/.config/packer/plugins/packer-plugin-testinfra
	@cp packer-plugin-testinfra ./provisioner/packer-plugin-testinfra
	@PACKER_ACC=1 go test -v ./provisioner/testinfra_acceptance_test.go -timeout=5m

install-packer-sdc:
	@go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@latest

#release-docs: install-packer-sdc
#	@packer-sdc renderdocs -src docs -partials docs-partials/ -dst docs/
#	@/bin/sh -c "[ -d docs ] && zip -r docs.zip docs/"

plugin-check: install-packer-sdc build
	export PATH="${PATH}:$(shell go env GOPATH)/bin"
	@packer-sdc plugin-check packer-plugin-testinfra

generate: install-packer-sdc
	export PATH="${PATH}:$(shell go env GOPATH)/bin"
	@go generate ./...
	#packer-sdc renderdocs -src ./docs -dst ./.docs -partials ./docs-partials

test: unit accept plugin-check
