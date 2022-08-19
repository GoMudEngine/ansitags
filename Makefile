.DEFAULT_GOAL := build

BIN ?= ansigo

.PHONY: test
test: ### Run tests with some verbose information filtered out and some minor styling
	@CGO_ENABLED=1 go test -v -race \
		-coverprofile=testdata/coverage_tests.out \
		. \
		| grep -v "=== RUN\|\d: PASS:" \
		| sed -E "s/^coverage: ([0-9\.%]*) of statements/`printf "\033[97;1mcoverage: \033[93;1m"`\1`printf "\033[0m \033[97;1mof statements\033[0m"`/" \
		| sed -E "s/^([[:space:]]*)--- PASS: (.*)/`printf "\033[96m"`\1--- PASS: \2`printf "\033[0m"`/" \
		| sed -E "s/^([[:space:]]*)([-]*)([[:space:]]*)FAIL(.*)/`printf "\033[91m"`\1\2\3FAIL\4`printf "\033[0m"`/"


.PHONY: coverage-html 
coverage-html: ### Saves a coverage html file
	@go tool cover -html=testdata/coverage_tests.out -o testdata/coverage_tests.html 

.PHONY: view-coverage 
view-coverage: coverage-html  # Opens/views coverage report in browser
	@open testdata/coverage_tests.html

.PHONY: build
build: 
	CGO_ENABLED=0 go build -trimpath -a -o example/bin/$(BIN) ./example