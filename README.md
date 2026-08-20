# Build
- Build dependencies: golang-go, valac, gcc, libsoup2.4-dev, libnotify-dev, libvte-2.91-dev, git
- Runtime dependencies: bleachbit, nyx, pkexec, sudo, tor, iptables, network-manager, psmisc, systemd, libglib2.0-0, libgtk-3-0, libnotify4, libsoup2.4-1, libvte-2.91-0
- Commands to build:
1. `git clone https://gitlab.com/parrotsec/packages/anonsurf/`
2. `cd anonsurf && make build && sudo make install`
3. `make check`

# Source layout
- `src/`: Go command-line tools, SafeKill, and system backend.
- `gui/`: Vala GTK interface.

- List of default bridges are at https://gitlab.torproject.org/tpo/anti-censorship/team/-/wikis/Default-Bridges
# Thank to
- Ranjeetsih, who wrote example of sys-tray with GTK at https://www.codeproject.com/Articles/27142/Minimize-to-tray-with-GTK
- Vivek Gite, who wrote tutorial of hook script for dhclient https://www.cyberciti.biz/faq/dhclient-etcresolvconf-hooks/