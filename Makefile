BINARY := bin/prism
SOURCES := $(shell find . -name '*.go' -not -path './fixture/*' -not -path './vendor/*')

.PHONY: build demo install clean record

$(BINARY): $(SOURCES) go.mod go.sum
	go build -o $(BINARY) .

build: $(BINARY)

install:
	go install .

# Run fixture tests and pipe through prism; 'true' swallows the non-zero exit.
demo: $(BINARY)
	@go test -json ./fixture/... 2>/dev/null | ./$(BINARY); true


# Record a demo GIF with vhs (https://github.com/charmbracelet/vhs)
record: $(BINARY)
	vhs demo.tape

clean:
	rm -rf bin/
