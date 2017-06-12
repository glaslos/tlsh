VERSION := v0.1.0
BUILDSTRING := $(shell git log --pretty=format:'%h' -n 1)
VERSIONSTRING := tlsh version $(VERSION)+$(BUILDSTRING)
BUILDDATE := $(shell date -u -Iseconds)

clean:
	rm -rf dist/

LDFLAGS := "-X \"main.VERSION=$(VERSIONSTRING)\" -X \"main.BUILDDATE=$(BUILDDATE)\""

tag:
	git tag $(VERSION)
	git push origin --tags

.PHONY: build_release
build_release: clean
	cd app; gox -arch="amd64" -os="windows darwin linux" -output="../dist/tlsh-{{.Arch}}-{{.OS}}" -ldflags=$(LDFLAGS)
