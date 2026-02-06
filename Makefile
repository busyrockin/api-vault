CGO_CFLAGS  := -I/opt/homebrew/opt/sqlcipher/include
CGO_LDFLAGS := -L/opt/homebrew/opt/sqlcipher/lib -lsqlcipher
CGO_ENABLED := 1
GO_ENV      := CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_ENABLED=$(CGO_ENABLED)

build:
	$(GO_ENV) go build -o api-vault .

run:
	$(GO_ENV) go run . $(ARGS)

test:
	$(GO_ENV) go test ./...

clean:
	rm -f api-vault

.PHONY: build run test clean
