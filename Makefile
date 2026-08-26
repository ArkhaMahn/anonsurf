GO ?= go
GO-SRC-DIR = src
GO-MODULE = gitlab.com/parrotsec/packages/anonsurf
GO-BUILD-FLAGS = -trimpath
GO-LDFLAGS = -s -w
GO-BUILD-MODE ?= module
GO-GOPATH ?= $(CURDIR)/build/gopath:/usr/share/gocode
GO-GOPATH-SRC = $(CURDIR)/build/gopath/src/$(GO-MODULE)
BUILD-DEST = build/release

VERSION ?= $(shell sed -n '1s/^[^(]*(\([^)]*\)).*/\1/p' debian/changelog)

ifeq ($(GO-BUILD-MODE),gopath)
GO-BUILD-DEPS = prepare-go-gopath
GO-BUILD-ENV = GO111MODULE=off GOPATH=$(GO-GOPATH)
GO-PKG-PREFIX = $(GO-MODULE)
else
GO-BUILD-DEPS =
GO-BUILD-ENV = GO111MODULE=on
GO-PKG-PREFIX = .
endif

.PHONY: build check clean prepare-go-gopath

all: build install

clean:
	rm -rf build

$(BUILD-DEST):
	mkdir -p $(BUILD-DEST)

prepare-go-gopath: $(BUILD-DEST)
	mkdir -p $(dir $(GO-GOPATH-SRC))
	rm -rf $(GO-GOPATH-SRC)
	cp -a $(CURDIR)/$(GO-SRC-DIR) $(GO-GOPATH-SRC)

anonsurf-cli: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS) -X main.version=$(VERSION)" -o ../$(BUILD-DEST)/anonsurf $(GO-PKG-PREFIX)/cmd/anonsurf-cli

dnstool: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS)" -o ../$(BUILD-DEST)/dnstool $(GO-PKG-PREFIX)/cmd/dnstool

make-torrc: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS)" -o ../$(BUILD-DEST)/make-torrc $(GO-PKG-PREFIX)/cmd/make-torrc

anondaemon: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS)" -o ../$(BUILD-DEST)/anondaemon $(GO-PKG-PREFIX)/cmd/anondaemon

safekill: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS)" -o ../$(BUILD-DEST)/safekill $(GO-PKG-PREFIX)/cmd/safekill

build: anonsurf-cli dnstool make-torrc anondaemon safekill

check: $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && files=$$(gofmt -l .) && \
		if [ -n "$$files" ]; then echo "not gofmt-clean: $$files"; exit 1; fi
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) vet $(GO-PKG-PREFIX)/...
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GO-PKG-PREFIX)/cmd/...

inst-configs:
	cp configs/bridges.txt $(DESTDIR)/etc/anonsurf/.
	cp configs/onion.pac $(DESTDIR)/etc/anonsurf/.
	cp configs/anonsurf.cfg $(DESTDIR)/etc/anonsurf/.

	cp sys-units/anonsurfd.service $(DESTDIR)/lib/systemd/system/anonsurfd.service
	cp sys-units/anonsurf-recovery.service $(DESTDIR)/lib/systemd/system/anonsurf-recovery.service
	cp sys-units/tor@default.service.d/anonsurf-recovery.conf $(DESTDIR)/lib/systemd/system/tor@default.service.d/anonsurf-recovery.conf
	ln -sfn ../anonsurf-recovery.service $(DESTDIR)/lib/systemd/system/multi-user.target.wants/anonsurf-recovery.service

inst-bins:
	cp build/release/anonsurf $(DESTDIR)/usr/bin/anonsurf
	cp build/release/dnstool $(DESTDIR)/usr/bin/dnstool
	cp build/release/make-torrc $(DESTDIR)/usr/lib/anonsurf/make-torrc
	cp build/release/anondaemon $(DESTDIR)/usr/lib/anonsurf/anondaemon
	cp build/release/safekill $(DESTDIR)/usr/lib/anonsurf/safekill

inst-perms:
	chmod +x $(DESTDIR)/usr/lib/anonsurf/anondaemon $(DESTDIR)/usr/lib/anonsurf/make-torrc $(DESTDIR)/usr/lib/anonsurf/safekill $(DESTDIR)/usr/bin/anonsurf
	chmod 600 $(DESTDIR)/etc/anonsurf/anonsurf.cfg

mk-folders:
	mkdir -p $(DESTDIR)/etc/anonsurf/
	mkdir -p $(DESTDIR)/usr/lib/anonsurf/
	mkdir -p $(DESTDIR)/usr/bin/
	mkdir -p $(DESTDIR)/lib/systemd/system/
	mkdir -p $(DESTDIR)/lib/systemd/system/multi-user.target.wants/
	mkdir -p $(DESTDIR)/lib/systemd/system/tor@default.service.d/

install: mk-folders inst-bins inst-configs inst-perms

uninstall:
	rm -rf /etc/anonsurf/
	rm -rf /usr/lib/anonsurf/
	rm /usr/bin/anonsurf
	rm /lib/systemd/system/anonsurfd.service
	rm /lib/systemd/system/anonsurf-recovery.service
	rm /lib/systemd/system/multi-user.target.wants/anonsurf-recovery.service
	rm /lib/systemd/system/tor@default.service.d/anonsurf-recovery.conf
