VALA-FLAGS = --pkg gtk+-3.0 # GTK3
VALA-FLAGS += --pkg libsoup-2.4 # Use libsoup to parse HTML, build-time requires libsoup2.4-dev
VALA-FLAGS += --pkg libnotify # Use libnotify to send notifications, build-time requires libnotify-dev
VALA-FLAGS += --vapidir /usr/share/vala/vapi/ --pkg vte-2.91 # Use libvte to create a terminal. build-time requires libvte-2.91-dev. Libvte binding is at /usr/share/vala/vapi/vte-2.91.vapi
VALA-FLAGS += -X -Wno-incompatible-pointer-types
VALA-FLAGS += -X -Wno-discarded-qualifiers

all:
	mkdir -p build/
	valac $(VALA-FLAGS) anonsurf_gtk.vala -o build/anonsurf-gtk \
	layouts/anonsurf_boot.vala layouts/anonsurf_controller.vala layouts/titlebar.vala \
	system_interactions/notifications.vala system_interactions/system_tray_icon.vala system_interactions/system_tray_menu.vala system_interactions/threads.vala \
	dialogs/about.vala dialogs/tor_controller.vala \
