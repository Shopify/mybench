BENCHMARKS := $(notdir $(shell find benchmarks -mindepth 1 -maxdepth 1 -type d))

.PHONY: $(BENCHMARKS) docs clean test

all: $(BENCHMARKS)

$(BENCHMARKS):
	mkdir -p build
	go build -o build/$@ github.com/Shopify/mybench/benchmarks/$@

docs:
	make -C docs html

clean:
	rm build -rf

test:
	go test -v ./...
