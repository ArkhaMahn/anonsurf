# Require libnotify-dev, libsoup2.4-dev, libvte-2.91-dev
# Libvte binding is at /usr/share/vala/vapi/vte-2.91.vapi
VALA-FLAGS = --pkg gtk+-3.0 --pkg libsoup-2.4 --pkg libnotify --vapidir /usr/share/vala/vapi/ --pkg vte-2.91 # --pkg gio-2.0 FIXME VTE not found

all:
	mkdir -p build/
	valac $(VALA-FLAGS) anonsurf_gtk.vala -o build/anonsurf-gtk \
	layouts/main_box.vala layouts/titlebar.vala layouts/area_boot_status.vala layouts/area_control_buttons.vala \
	systray/systray.vala \
	notifications/notify.vala \
	dialogs/about.vala dialogs/tor_controller.vala \
	threads/threads.vala