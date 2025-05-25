.PHONY: generate build clean run deps

deps:
	@command -v clang >/dev/null 2>&1 || (echo "clang not found. Please install: apt install clang" && exit 1)
	@command -v llvm-strip >/dev/null 2>&1 || (echo "llvm-strip not found. Please install: apt install llvm" && exit 1)

generate: deps
	go generate .

build: generate
	go build -o sniff-writes .

clean:
	rm -f sniff-writes
	rm -f sniffwrites_bpfel.go sniffwrites_bpfeb.go sniffwrites_bpfel.o sniffwrites_bpfeb.o

run: build
	sudo ./sniff-writes monitor

test: build
	@echo "Testing the program (requires sudo)..."
	@echo "This will monitor for 5 seconds then exit"
	timeout 5s sudo ./sniff-writes monitor -v || true

demo: build
	@echo "Demo: Monitor /tmp directory with table format for 10 seconds"
	timeout 10s sudo ./sniff-writes monitor -d /tmp -f table --show-fd -v || true

help:
	@echo "Available targets:"
	@echo "  deps     - Check dependencies"
	@echo "  generate - Generate eBPF Go bindings"
	@echo "  build    - Build the program"
	@echo "  clean    - Clean generated files"
	@echo "  run      - Build and run with sudo"
	@echo "  test     - Quick test run (5 seconds)"
	@echo "  demo     - Demo with /tmp monitoring (10 seconds)"