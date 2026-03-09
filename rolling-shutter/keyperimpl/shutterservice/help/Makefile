# Makefile for Emitter.sol

.PHONY: all clean test

# Variables
SOLC := solc
ABIGEN := abigen
CONTRACT := Emitter

# Default target: generate Go bindings
all: $(CONTRACT).go

# Compile Solidity to ABI and Binary
$(CONTRACT).json: $(CONTRACT).sol
	@echo "Compiling Solidity contract..."
	$(SOLC) --combined-json abi,bin $(CONTRACT).sol > $(CONTRACT).json

# Generate Go bindings using abigen
$(CONTRACT).go: $(CONTRACT).json
	@echo "Generating Go bindings..."
	$(ABIGEN) \
		--combined-json $(CONTRACT).json \
		--pkg shutterservice \
		--out ../$(CONTRACT).go

# Clean generated files
clean:
	rm -f $(CONTRACT).json ../$(CONTRACT).go

test:
	go test -run TestMyEVM
