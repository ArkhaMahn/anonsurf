VALA-FLAGS = --pkg gtk+-3.0 # GTK3
VALA-FLAGS += --pkg libsoup-2.4 # Use libsoup to parse HTML, build-time requires libsoup2.4-dev
VALA-FLAGS += --pkg libnotify # Use libnotify to send notifications, build-time requires libnotify-dev
VALA-FLAGS += --vapidir /usr/share/vala/vapi/ --pkg vte-2.91 # Use libvte to create a terminal. build-time requires libvte-2.91-dev. Libvte binding is at /usr/share/vala/vapi/vte-2.91.vapi
VALA-FLAGS += -X -Wno-incompatible-pointer-types # Suppress Vala's warning. Check https://stackoverflow.com/a/66559047
VALA-FLAGS += -X -Wno-discarded-qualifiers # Supress Vala's warning. Check https://stackoverflow.com/a/66559047

VALA-SRC-DIR = src/anonsurf-gtk
VALA-SRC-LAYOUTS = $(VALA-SRC-DIR)/layouts
VALA-SRC-DIALOGS = $(VALA-SRC-DIR)/dialogs
VALA-SRC-SYS = $(VALA-SRC-DIR)/system_interactions

NIM-FLAGS = c --nimcache:/tmp/tmpcache -d:ssl -d:release

.PHONY: build

all: build install

clean:
	rm -rf build

anonsurf-gtk:
	valac $(VALA-FLAGS) $(VALA-SRC-DIR)/anonsurf_gtk.vala \
	$(VALA-SRC-LAYOUTS)/anonsurf_boot.vala $(VALA-SRC-LAYOUTS)/anonsurf_controller.vala $(VALA-SRC-LAYOUTS)/titlebar.vala \
	$(VALA-SRC-DIALOGS)/about.vala $(VALA-SRC-DIALOGS)/tor_controller.vala \
	$(VALA-SRC-SYS)/notifications.vala $(VALA-SRC-SYS)/system_tray_icon.vala $(VALA-SRC-SYS)/system_tray_menu.vala $(VALA-SRC-SYS)/threads.vala \
	-o build/release/anonsurf-gtk

anonsurf-cli:
	nim $(NIM-FLAGS) --out:build/release/anonsurf src/anonsurf_cli/anonsurf_cli.nim

dnstool:
	nim $(NIM-FLAGS) --out:build/release/dnstool src/dnstool/dnstool.nim

make-torrc:
	nim $(NIM-FLAGS) --out:build/release/make-torrc src/make_torrc/make_torrc.nim

build: anonsurf-cli anonsurf-gtk dnstool make-torrc

inst-launchers:
	# Copy launchers
	os_name = "$(shell cat /etc/os-release | grep "^ID=" | cut -d = -f 2)"

	if [ os_name = "parrot" ]; then \
		cp launchers/parrot/*.desktop $(DESTDIR)/usr/share/applications/; \
	else \
		cp launchers/debian/*.desktop $(DESTDIR)/usr/share/applications/; \
	fi

inst-configs:
	# Copy configs
	cp configs/bridges.txt $(DESTDIR)/etc/anonsurf/.
	cp configs/onion.pac $(DESTDIR)/etc/anonsurf/.

	# Copy daemon service
	cp sys-units/anonsurfd.service $(DESTDIR)/lib/systemd/system/anonsurfd.service

inst-bins:
	# Copy binaries to system
	cp build/release/anonsurf $(DESTDIR)/usr/bin/anonsurf
	cp build/release/anonsurf-gtk $(DESTDIR)/usr/bin/anonsurf-gtk
	cp build/release/dnstool $(DESTDIR)/usr/bin/dnstool
	cp build/release/make-torrc $(DESTDIR)/usr/lib/anonsurf/make-torrc
	cp scripts/* $(DESTDIR)/usr/lib/anonsurf/

mk-folders:
	# Create all folders
	mkdir -p $(DESTDIR)/etc/anonsurf/
	mkdir -p $(DESTDIR)/usr/lib/anonsurf/
	mkdir -p $(DESTDIR)/usr/bin/
	mkdir -p $(DESTDIR)/usr/share/applications/
	mkdir -p $(DESTDIR)/lib/systemd/system/

install: mk-folders inst-bins inst-configs inst-launchers

uninstall:
	# Remove config files
	rm -rf /etc/anonsurf/
	# Remove daemon scripts and some other binaries
	rm -rf /usr/lib/anonsurf/
	# Remove binaries
	rm /usr/bin/anonsurf
	rm /usr/bin/anonsurf-gtk
	# Remove systemd unit
	rm /lib/systemd/system/anonsurfd.service
	# Remove launchers
	rm /usr/share/applications/anonsurf*.desktop
