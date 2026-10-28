# Maintainer: AnonSurf CachyOS port
pkgname=anonsurf
pkgver=6.0.1
pkgrel=4
pkgdesc='Anonymization toolkit that routes supported system traffic through Tor'
arch=('x86_64' 'aarch64')
url='https://www.parrotsec.org/'
license=('GPL-2.0-only')
# Older experimental builds shipped the stack split across these names.
conflicts=('anonsurf-gui' 'anonsurf-cli')
replaces=('anonsurf-gui' 'anonsurf-cli')
depends=(
  'iptables'
  'networkmanager'
  'procps-ng'
  'tor'
  'systemd'
  'sudo'
  'nyx'
  'konsole'
  'inter-font'
  'qt6-base'
  'qt6-declarative'
)
optdepends=(
  'bleachbit: clean application caches when starting AnonSurf'
  'tor-geoipdb: GeoIP database for Tor'
  'obfs4proxy: obfs4 bridges'
  'snowflake: Snowflake bridges'
  'qt6-wayland: native Wayland rendering for the Qt 6 interface'
)
makedepends=(
  'cmake'
  'gcc'
  'go'
  'pkgconf'
  'qt6-base'
  'qt6-declarative'
  'qt6-tools'
)
backup=(
  'etc/anonsurf/anonsurf.cfg'
  'etc/anonsurf/bridges.txt'
  'etc/anonsurf/onion.pac'
)
install=anonsurf.install
source=("$pkgname-$pkgver.tar.gz")
sha256sums=('38aabb226d4c1bd1648b8dccda5a6e950b8f7b98aa29744aea22af04b1591ace')

build() {
  cd "$pkgname-$pkgver"
  export CGO_ENABLED=0
  make -f Makefile.arch VERSION="$pkgver" build
}

check() {
  cd "$pkgname-$pkgver"
  export CGO_ENABLED=0
  make -f Makefile.arch VERSION="$pkgver" check
}

package() {
  cd "$pkgname-$pkgver"
  make -f Makefile.arch VERSION="$pkgver" DESTDIR="$pkgdir" install

  install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
