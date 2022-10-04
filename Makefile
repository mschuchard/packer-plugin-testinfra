.PHONY: build

tidy:
	@go mod tidy

build: tidy
	@go build -o packer-plugin-testinfra

unit: tidy
	@go test -v

acceptance: build
	# vagrant up and vagrant suspend one-time in fixtures dir
	@mkdir -p ~/.packer.d/plugins/
	# current packer-sdk bug cannot find plugin installed locally
	@cp packer-plugin-testinfra ~/.packer.d/plugins/packer-plugin-testinfra
	@PACKER_ACC=1 go test -v packer-provisioner-testinfra_acceptance_test.go -timeout=5m

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

test: unit acceptance plugin-check
