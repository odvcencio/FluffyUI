.PHONY: bench bench-save test build vet

build:
	go build ./...

test:
	go test ./... -count=1 -timeout=300s

vet:
	go vet ./...

bench:
	go test ./state/ -bench=. -benchmem -count=3 -timeout=120s
	go test ./style/ -bench=. -benchmem -count=3 -timeout=120s
	go test ./compositor/ -bench=. -benchmem -count=3 -timeout=120s
	go test ./scroll/ -bench=. -benchmem -count=3 -timeout=120s
	go test ./runtime/ -bench=. -benchmem -count=3 -timeout=120s
	go test ./widgets/ -bench=. -benchmem -count=3 -timeout=120s

bench-save:
	go test ./... -bench=. -benchmem -count=5 -timeout=300s | tee benchmarks.txt
