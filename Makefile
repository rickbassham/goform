GO_FOLDERS=./...

pre-commit: go-test go-lint go-mod-tidy go-doc

commit:
	@git cz

setup:
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@echo "#!/bin/bash\ncat \$$1 | commitlint" > .git/hooks/commit-msg
	@chmod +x .git/hooks/commit-msg

utilities:
	@go get -u github.com/robertkrimen/godocdown/godocdown

	@npm install -g eslint

	@npm install -g \
	commitizen \
	cz-conventional-changelog \
	@commitlint/prompt-cli \
	@commitlint/config-conventional

	@npm install -g \
	semantic-release \
	@semantic-release/commit-analyzer \
	@semantic-release/release-notes-generator \
	@semantic-release/changelog \
	@semantic-release/git

go-test:
	@go test $(GO_FOLDERS)

go-lint:
	@golangci-lint run \
	--exclude-use-default=false --disable-all \
	--enable golint --enable gosec --enable interfacer --enable unconvert \
	--enable goimports --enable goconst --enable gocyclo --enable misspell \
	--enable scopelint \
	$(GO_FOLDERS)

go-mod-tidy:
	@go mod tidy

go-doc:
	@godocdown > ./README.md
