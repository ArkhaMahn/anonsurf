VALA-FLAGS = --pkg gtk+-3.0
VALA-FLAGS += --pkg libnotify
VALA-FLAGS += --vapidir /usr/share/vala/vapi/ --pkg vte-2.91
VALA-FLAGS += -X -Wno-incompatible-pointer-types

VALA-SRC-DIR = gui/anonsurf-gtk
VALA-SRC-CONTROLS = $(VALA-SRC-DIR)/controls
VALA-SRC-LAYOUTS = $(VALA-SRC-DIR)/layouts
VALA-SRC-DIALOGS = $(VALA-SRC-DIR)/dialogs
VALA-SRC-SYS = $(VALA-SRC-DIR)/system_interactions

GO ?= go
GO-SRC-DIR = src
GO-MODULE = gitlab.com/parrotsec/packages/anonsurf
GO-BUILD-FLAGS = -trimpath
GO-LDFLAGS = -s -w
GO-BUILD-MODE ?= module
GO-GOPATH ?= $(CURDIR)/build/gopath:/usr/share/gocode
GO-GOPATH-SRC = $(CURDIR)/build/gopath/src/$(GO-MODULE)
BUILD-DEST = build/release

ifeq ($(GO-BUILD-MODE),gopath)
GO-BUILD-DEPS = prepare-go-gopath
GO-BUILD-ENV = GO111MODULE=off GOPATH=$(GO-GOPATH)
GO-PKG-PREFIX = $(GO-MODULE)
else
GO-BUILD-DEPS =
GO-BUILD-ENV = GO111MODULE=on
GO-PKG-PREFIX = .
endif

os_name := $(shell sed -n 's/^ID=//p' /etc/os-release 2>/dev/null | tr -d '"')

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

anonsurf-gtk: $(BUILD-DEST)
	valac $(VALA-FLAGS) $(VALA-SRC-DIR)/anonsurf_gtk.vala \
	$(VALA-SRC-LAYOUTS)/box_boot.vala $(VALA-SRC-LAYOUTS)/box_controller.vala $(VALA-SRC-LAYOUTS)/titlebar.vala \
	$(VALA-SRC-CONTROLS)/network_buttons.vala $(VALA-SRC-CONTROLS)/tor_buttons.vala \
	$(VALA-SRC-DIALOGS)/about.vala $(VALA-SRC-DIALOGS)/tor_controller.vala $(VALA-SRC-DIALOGS)/ask_kill.vala \
	$(VALA-SRC-SYS)/notifications.vala $(VALA-SRC-SYS)/system_tray_icon.vala $(VALA-SRC-SYS)/system_tray_menu.vala $(VALA-SRC-SYS)/threads.vala \
	-o $(BUILD-DEST)/anonsurf-gtk

anonsurf-cli: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS)" -o ../$(BUILD-DEST)/anonsurf $(GO-PKG-PREFIX)/cmd/anonsurf-cli

dnstool: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS)" -o ../$(BUILD-DEST)/dnstool $(GO-PKG-PREFIX)/cmd/dnstool

make-torrc: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS)" -o ../$(BUILD-DEST)/make-torrc $(GO-PKG-PREFIX)/cmd/make-torrc

anondaemon: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS)" -o ../$(BUILD-DEST)/anondaemon $(GO-PKG-PREFIX)/cmd/anondaemon

safekill: $(BUILD-DEST) $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) build $(GO-BUILD-FLAGS) -ldflags="$(GO-LDFLAGS)" -o ../$(BUILD-DEST)/safekill $(GO-PKG-PREFIX)/cmd/safekill

build: anonsurf-cli anonsurf-gtk dnstool make-torrc anondaemon safekill

check: $(GO-BUILD-DEPS)
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) $(GO) vet $(GO-PKG-PREFIX)/...
	cd $(GO-SRC-DIR) && $(GO-BUILD-ENV) GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GO-PKG-PREFIX)/cmd/...

inst-launchers:
	if [ "$(os_name)" = "parrot" ]; then \
		cp launchers/parrot/*.desktop $(DESTDIR)/usr/share/applications/; \
	else \
		cp launchers/debian/*.desktop $(DESTDIR)/usr/share/applications/; \
	fi

inst-configs:
	cp configs/bridges.txt $(DESTDIR)/etc/anonsurf/.
	cp configs/onion.pac $(DESTDIR)/etc/anonsurf/.
	cp configs/anonsurf.cfg $(DESTDIR)/etc/anonsurf/.

	cp sys-units/anonsurfd.service $(DESTDIR)/lib/systemd/system/anonsurfd.service

inst-bins:
	cp build/release/anonsurf $(DESTDIR)/usr/bin/anonsurf
	cp build/release/anonsurf-gtk $(DESTDIR)/usr/bin/anonsurf-gtk
	cp build/release/dnstool $(DESTDIR)/usr/bin/dnstool
	cp build/release/make-torrc $(DESTDIR)/usr/lib/anonsurf/make-torrc
	cp build/release/anondaemon $(DESTDIR)/usr/lib/anonsurf/anondaemon
	cp build/release/safekill $(DESTDIR)/usr/lib/anonsurf/safekill

inst-perms:
	chmod +x $(DESTDIR)/usr/lib/anonsurf/anondaemon $(DESTDIR)/usr/lib/anonsurf/make-torrc $(DESTDIR)/usr/lib/anonsurf/safekill $(DESTDIR)/usr/bin/anonsurf $(DESTDIR)/usr/bin/anonsurf-gtk

inst-icons:
	cp icons/*.png $(DESTDIR)/usr/share/icons/.

mk-folders:
	mkdir -p $(DESTDIR)/etc/anonsurf/
	mkdir -p $(DESTDIR)/usr/lib/anonsurf/
	mkdir -p $(DESTDIR)/usr/bin/
	mkdir -p $(DESTDIR)/usr/share/applications/
	mkdir -p $(DESTDIR)/usr/share/icons/
	mkdir -p $(DESTDIR)/lib/systemd/system/

install: mk-folders inst-bins inst-configs inst-launchers inst-icons inst-perms

uninstall:
	rm -rf /etc/anonsurf/
	rm -rf /usr/lib/anonsurf/
	rm /usr/bin/anonsurf
	rm /usr/bin/anonsurf-gtk
	rm /lib/systemd/system/anonsurfd.service
	rm /usr/share/applications/anonsurf*.desktop
