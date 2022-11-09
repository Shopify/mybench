BENCHMARKS := $(notdir $(shell find benchmarks -mindepth 1 -maxdepth 1 -type d))

.PHONY: $(BENCHMARKS) clean test

all: $(BENCHMARKS)

$(BENCHMARKS):
	mkdir -p build
	go build -o build/$@ github.com/Shopify/mybench/benchmarks/$@

clean:
	rm build -rf

test:
	go test -count=1 -test.v
