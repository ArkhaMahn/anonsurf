all:
	mkdir -p build/
	# require libsoup2.4-dev
	valac --pkg gtk+-3.0 --pkg libsoup-2.4 anonsurf_gtk.vala -o build/anonsurf-gtk \
	layouts/main_box.vala layouts/titlebar.vala \
	systray/systray.vala \
	dialogs/about.vala  dialogs/options.vala