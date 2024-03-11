
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

build:
	rm -rf tmp/extensions tmp/uploads
	pnpm install
	go get github.com/mattn/goreman@latest
	mkdir -p dist
	mkdir -p tmp/extensions/directus-extension-call-go
	ln -s $(PWD)/dist $(PWD)/tmp/extensions/directus-extension-call-go/dist
	ln -s $(PWD)/package.json $(PWD)/tmp/extensions/directus-extension-call-go/package.json

data:
	@pnpm data

lint:
	@pnpm lint
	go install ./...
	go vet ./...
	linter ./...

serve:
	@goreman -set-ports=false start

gofmt:
	@gofmt -s -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)
	@impsort . -p github.com/altipla-consulting/directus-call-go

test:
	go test -v -race ./...
