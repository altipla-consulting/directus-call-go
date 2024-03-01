
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

build:
	@pnpm install

data:
	@pnpm data

lint:
	@pnpm lint
	go install ./...
	go vet ./...
	linter ./...

serve.ext:
	@docker compose up -d directus
	@pnpm dev

serve.target:
	@reloader run -r ./testapp -w ./callgo

gofmt:
	@gofmt -s -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)
	@impsort . -p github.com/altipla-consulting/directus-call-go

test:
	go test -v -race ./...
