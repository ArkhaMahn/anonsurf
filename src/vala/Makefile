# Require libnotify-dev, libsoup2.4-dev
VALA-FLAGS = --pkg gtk+-3.0 --pkg libsoup-2.4 --pkg libnotify # --pkg gio-2.0

all:
	mkdir -p build/
	valac $(VALA-FLAGS) anonsurf_gtk.vala -o build/anonsurf-gtk \
	layouts/main_box.vala layouts/titlebar.vala \
	systray/systray.vala \
	notifications/notify.vala \
	dialogs/about.vala \
	threads/threads.vala