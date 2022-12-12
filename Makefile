VERSION := v0.3.0
NAME := tlsh
BUILDSTRING := $(shell git log --pretty=format:'%h' -n 1)
VERSIONSTRING := $(NAME) version $(VERSION)+$(BUILDSTRING)
BUILDDATE := $(shell date -u -Iseconds)
OUTPUT = dist/$(NAME)
LDFLAGS := "-X \"main.VERSION=$(VERSIONSTRING)\" -X \"main.BUILDDATE=$(BUILDDATE)\""

default: build

build: $(OUTPUT)

$(OUTPUT): app/tlsh.go pearson.go tlsh.go
	@mkdir -p dist/
	go build -o $(OUTPUT) -ldflags=$(LDFLAGS) app/tlsh.go

.PHONY: clean
clean:
	rm -rf dist/

.PHONY: tag
tag:
	git tag $(VERSION)
	git push origin --tags

.PHONY: build_release
build_release: clean
	cd app; gox -arch="amd64" -os="windows darwin linux" -output="../dist/$(NAME)-{{.Arch}}-{{.OS}}" -ldflags=$(LDFLAGS)

.PHONY: bench
bench:
	go test -bench=.

.PHONY: test
test:
	go test ./...

.PHONY: profile
profile:
	@mkdir -p pprof/
	go test -cpuprofile pprof/cpu.prof -memprofile pprof/mem.prof -bench .
	go tool pprof -pdf pprof/cpu.prof > pprof/cpu.pdf
	xdg-open pprof/cpu.pdf
	go tool pprof -weblist=.* pprof/cpu.prof

.PHONY: benchcmp
benchcmp:
	# ensure no govenor weirdness
	# sudo cpufreq-set -g performance
	go test -test.benchmem=true -run=NONE -bench=. ./... > bench_current.test
	git stash save "stashing for benchcmp"
	@go test -test.benchmem=true -run=NONE -bench=. ./... > bench_head.test
	git stash pop
	benchstat bench_head.test bench_current.test

.PHONY: benchbcmp
benchbcmp:
	# ensure no govenor weirdness
	# sudo cpufreq-set -g performance
	go test -test.benchmem=true -run=NONE -bench=. ./... > bench_current.test
	git stash save "stashing for benchcmp"
	git checkout main
	@go test -test.benchmem=true -run=NONE -bench=. ./... > bench_main.test
	git checkout -
	git stash pop
	benchstat bench_main.test bench_current.test
