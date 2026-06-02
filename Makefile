BINARY := bin/prism
SOURCES := $(shell find . -name '*.go' -not -path './fixture/*' -not -path './vendor/*')

.PHONY: build demo demo-race demo-bench install clean record

$(BINARY): $(SOURCES) go.mod go.sum
	go build -o $(BINARY) .

build: $(BINARY)

install:
	go install .

# Run fixture tests and pipe through prism; 'true' swallows the non-zero exit.
demo: $(BINARY)
	@go test -json ./fixture/... 2>/dev/null | ./$(BINARY); true


# Show the race detector RACE card.
demo-race: $(BINARY)
	@go test -race -json ./fixture/race/... 2>/dev/null | ./$(BINARY); true

# Show the benchmark table (styled panel + copyable markdown).
# -run='^$$' skips tests so the benchmark phase runs (go test skips benchmarks
# when tests fail, and the calc fixture fails on purpose).
demo-bench: $(BINARY)
	@go test -bench=. -benchmem -run='^$$' -json ./fixture/calc/... 2>/dev/null | ./$(BINARY); true

# Record a demo GIF with vhs (https://github.com/charmbracelet/vhs)
record: $(BINARY)
	vhs demo.tape

clean:
	rm -rf bin/
