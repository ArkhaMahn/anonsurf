# Build
- Build dependencies: golang-go, git
- Runtime dependencies: bleachbit, nyx, sudo, tor, iptables, network-manager, procps, systemd
- Commands to build:
1. `git clone https://gitlab.com/parrotsec/packages/anonsurf/`
2. `cd anonsurf && make build && sudo make install`
3. `make check`

# Source layout
- `src/cmd/`: Go command-line entry points.
- `src/internal/anonsurf/`: lifecycle orchestration and transactional recovery.
- `src/internal/dns/`, `firewall/`, `ipv6/`, `tor/`: network domains.
- `src/internal/command/`, `fileops/`, `systemd/`: operating-system adapters.
- `icons/`: AnonSurf and Nyx logos retained for the future interface.

Interrupted sessions are recovered automatically at boot before Tor starts.

- List of default bridges are at https://gitlab.torproject.org/tpo/anti-censorship/team/-/wikis/Default-Bridges
# Thank to
- Vivek Gite, who wrote tutorial of hook script for dhclient https://www.cyberciti.biz/faq/dhclient-etcresolvconf-hooks/