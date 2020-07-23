all:

.PHONY: all
all: lint all-binaries

.PHONY: all-binaries
all-binaries: $(patsubst cmd/%,build/%,$(wildcard cmd/*))

.PHONY: generate-go
generate-go:
	go generate ./...

build/%: APP_NAME = $*
build/%: cmd/% generate-go
	@mkdir -p build
	go build -o $@ ./$<

# These are escape newlines, looks super weird.
lint: \
	lint-go-deadcode \
	lint-go-misspell \
	lint-go-vet \
	lint-go-nakedret

lint-go-deadcode:
	 deadcode ./

lint-go-misspell:
	misspell -error -f 'misspell: {{ .Filename }}:{{ .Line }}:{{ .Column }}:corrected {{ printf "%q" .Original }} to {{ printf "%q" .Corrected }}' ./

lint-go-vet:
	go vet ./...

lint-go-nakedret:
	nakedret ./...
