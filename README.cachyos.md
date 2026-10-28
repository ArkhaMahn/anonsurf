# AnonSurf on CachyOS / Arch Linux

This tree contains a CachyOS/Arch port of [ParrotSec AnonSurf](https://github.com/ParrotSec/anonsurf).
It builds the complete 6.0 stack: the Go daemon, CLI, DNS helper, SafeKill and
the Qt 6 / QML desktop interface.

## What was adapted for Arch

AnonSurf was written for Parrot's Debian packaging. The following changes make
it work on Arch-based systems:

| Area | Parrot / Debian | CachyOS / Arch |
| --- | --- | --- |
| Tor unit | `tor@default.service` | `tor.service` (detected at runtime) |
| Tor drop-in | `tor@default.service.d/` | `tor.service.d/` |
| Systemd units | `/lib/systemd/system` | `/usr/lib/systemd/system` |
| Generated `torrc` | `RunAsDaemon 1`, cookie in `/run` | `RunAsDaemon 0`, `User tor`, cookie in `/var/lib/tor` |
| Firewall owner rule | `debian-tor` | `tor` (resolved at runtime) |
| System DNS | nss-dns via `/etc/resolv.conf` | `systemd-resolved` suspended and restored around routing |
| GUI font | `fonts-inter` | `inter-font` |
| Qt QML modules | distribution specific | `qt6-declarative` |

The runtime detection lives in `src/internal/platform/platform.go` and
`src/internal/tor/platform.go`, so the same binary keeps working on Parrot and
on Arch.

Arch's `tor.service` runs Tor in the foreground under a hardened sandbox that
only allows writes to `/var/lib/tor` and `/var/log/tor`. Without the
adaptations above Tor either daemonizes away from systemd or cannot create its
control cookie, which breaks `anonsurf start` and `anonsurf changeid`.

### systemd-resolved

CachyOS resolves names through `systemd-resolved` and `nss-resolve`, which
answers glibc lookups directly and ignores `/etc/resolv.conf`. Writing Tor's
DNSPort into `resolv.conf` alone would therefore not take effect.

When routing starts, AnonSurf now:

1. writes a NetworkManager `dns=none` hook and reloads NetworkManager, so it
   stops pushing DNS state to resolved;
2. records that resolved was active in
   `/var/lib/anonsurf/systemd-resolved.was-active`, runtime-masks
   `systemd-resolved.service` and stops it. The mask is needed because the
   service is socket- and D-Bus-activated and would otherwise be started again
   by the next lookup;
3. points `/etc/resolv.conf` at Tor's DNSPort.

`anonsurf stop`, a failed start and `anonsurf-recovery.service` restore the
original `resolv.conf`, network settings and restart resolved (clearing the
runtime mask first). If AnonSurf is interrupted, the marker file makes the
recovery path restart the service on the next boot. The firewall's UDP/53
redirect to Tor remains as a safety net even if resolved is restarted on
demand.

## Build and install (recommended)

Build natively on CachyOS so the binaries link against your installed glibc and
Qt 6:

```bash
sudo pacman -S --needed base-devel cmake go qt6-base qt6-declarative qt6-tools
makepkg -si
```

`makepkg` uses the included `PKGBUILD`, verifies the source tarball, builds
everything with `Makefile.arch` and installs the `anonsurf` package.

## Install a prebuilt package

A package built from this tree is available as
`anonsurf-6.0.1-4-x86_64.pkg.tar.zst`:

```bash
sudo pacman -U anonsurf-6.0.1-4-x86_64.pkg.tar.zst
```

SHA-256: `9101156e03e90a9e2f3ebb5352caad4606b906ce31ec092a7a88cc09f5166492`

Prefer the native build when possible. The prebuilt package was assembled on a
non-Arch system, so it can lag behind the glibc/Qt versions of your CachyOS
installation.

If an earlier broken package is already installed, remove it first so no stale
files remain:

```bash
sudo pacman -Rns anonsurf anonsurf-gui anonsurf-cli 2>/dev/null
sudo rm -f /usr/share/applications/org.parrotsec.AnonsurfCli.desktop
sudo update-desktop-database -q /usr/share/applications 2>/dev/null
```

The package now also declares the old `anonsurf-gui`/`anonsurf-cli` names as
conflicts, and its install hook removes leftover launchers from earlier
experimental builds.

## Usage

```bash
anonsurf status          # current state
sudo anonsurf start      # route traffic through Tor
sudo anonsurf stop       # restore the original network
sudo anonsurf changeid   # request a new Tor identity
anonsurf ip              # check the current exit IP
anonsurf-gui             # Qt 6 desktop interface
```

The GUI is also available as **AnonSurf GUI** in the application menu. It runs
unprivileged and asks systemd (through polkit) for the privileged actions.

The command-line interface is terminal-only on this package: it installs no
separate menu entry or launcher, only the `anonsurf` command.

The package enables `anonsurf-recovery.service` so an interrupted session is
rolled back on the next boot. Routing itself is not enabled until you start it.

## Dependencies

Runtime dependencies installed automatically:

`iptables`, `networkmanager`, `procps-ng`, `tor`, `systemd`, `sudo`, `nyx`,
`konsole`, `inter-font`, `qt6-base`, `qt6-declarative`

Optional extras:

* `bleachbit` – clear application caches on start
* `tor-geoipdb` – GeoIP database
* `obfs4proxy` / `snowflake` – censorship circumvention bridges
* `qt6-wayland` – native Wayland rendering for the GUI

## Troubleshooting

**The AnonSurf GUI entry does nothing.** Run it from a terminal so Qt can print
its error:

```bash
anonsurf-gui
```

If the command is missing, an old split package is probably still installed:

```bash
pacman -Q anonsurf anonsurf-gui anonsurf-cli
pacman -Qo /usr/bin/anonsurf-gui
```

Remove stale packages and launchers, then reinstall:

```bash
sudo pacman -Rns anonsurf anonsurf-gui anonsurf-cli 2>/dev/null
sudo rm -f /usr/share/applications/org.parrotsec.AnonsurfCli.desktop
sudo pacman -U anonsurf-6.0.1-4-x86_64.pkg.tar.zst
```

If an old GUI process is still running, quit it from the tray or run
`pkill anonsurf-gui` once; the current build no longer exits silently when
activation of an existing instance fails.

**The GUI exits with code 127 and mentions a copy relocation.**

```
warning: copy relocation against non-copyable protected symbol
`_ZN6QTimer16staticMetaObjectE' in `/usr/lib/libQt6Core.so.6'
```

Older packages put `R_X86_64_COPY` relocations for Qt data symbols into the
executable. Arch's Qt 6.12 marks its libraries with
`GNU_PROPERTY_1_NEEDED_INDIRECT_EXTERN_ACCESS`, and glibc refuses to resolve
those relocations. The GUI is now built as fully position-independent and
linked with `-Wl,-z,nocopyreloc`, so the binary contains no copy relocations.
Reinstall `anonsurf-6.0.1-4` or newer; check with:

```bash
readelf -rW /usr/bin/anonsurf-gui | grep R_X86_64_COPY
```

The command should print nothing.

**A CLI launcher is still in the menu.** Only `org.parrotsec.Anonsurf.desktop`
is shipped now. The package install hook deletes
`org.parrotsec.AnonsurfCli.desktop`; log out and back in if the desktop
environment still shows the cached entry.

## Notes

* AnonSurf routes supported TCP traffic and DNS lookups, not arbitrary UDP.
* Existing direct connections are blocked when routing starts; reconnect them.
* `sudo anonsurf stop` restores DNS, `systemd-resolved`, IPv6 and firewall state.
