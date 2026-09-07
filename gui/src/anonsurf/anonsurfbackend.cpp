#include "anonsurf/anonsurfbackend.h"

namespace {
constexpr auto IdentityUnit = "anonsurf-identity.service";
} // namespace

AnonSurfBackend::AnonSurfBackend(QObject *parent) : QObject(parent) {
  connect(&m_commands, &CommandClient::statusReady, this,
          &AnonSurfBackend::handleStatus);
  connect(&m_commands, &CommandClient::statusFailed, this,
          [this](const QString &message) {
            m_statusAvailable = false;
            m_protectionDetail = message;
            m_dnsMode = "Unavailable";
            m_dnsServers = "No data";
            invalidateIPCheck();
            emit changed();
          });
  connect(&m_commands, &CommandClient::ipReady, this,
          [this](const QString &address, bool isTor) {
            if (m_ipCheckStale) {
              finishAction("ip", false,
                           "Service or DNS status changed during the IP check. "
                           "Run it again.");
              return;
            }
            m_publicIP = address;
            m_torCheck = isTor ? "TOR" : "CLEARNET";
            finishAction("ip", true, {});
          });
  connect(
      &m_commands, &CommandClient::ipFailed, this,
      [this](const QString &message) { finishAction("ip", false, message); });
  connect(&m_commands, &CommandClient::monitorFailed, this,
          &AnonSurfBackend::setError);
  connect(&m_commands, &CommandClient::safeKillFinished, this,
          [this](bool success, const QString &message) {
            handleAction("safekill", success, message);
          });

  connect(&m_systemd, &SystemdClient::actionFinished, this,
          &AnonSurfBackend::handleAction);
  connect(&m_systemd, &SystemdClient::bootStateReady, this,
          [this](bool enabled) {
            m_bootEnabled = enabled;
            emit changed();
          });
  connect(&m_systemd, &SystemdClient::statusTextReady, this,
          [this](const QString &status) {
            if (m_busy && !status.isEmpty() && status != m_progress) {
              m_progress = status;
              emit changed();
            }
          });

  m_progressTimer.setInterval(750);
  connect(&m_progressTimer, &QTimer::timeout, &m_systemd,
          &SystemdClient::requestStatusText);
}

bool AnonSurfBackend::statusAvailable() const { return m_statusAvailable; }
bool AnonSurfBackend::anonsurfActive() const { return m_anonsurfActive; }
bool AnonSurfBackend::torActive() const { return m_torActive; }
bool AnonSurfBackend::bootEnabled() const { return m_bootEnabled; }
bool AnonSurfBackend::busy() const { return m_busy; }
QString AnonSurfBackend::protectionState() const { return m_protectionState; }
QString AnonSurfBackend::protectionDetail() const { return m_protectionDetail; }
QString AnonSurfBackend::dnsMode() const { return m_dnsMode; }
QString AnonSurfBackend::dnsServers() const { return m_dnsServers; }
QString AnonSurfBackend::publicIP() const { return m_publicIP; }
QString AnonSurfBackend::torCheck() const { return m_torCheck; }
QString AnonSurfBackend::progress() const { return m_progress; }
QString AnonSurfBackend::errorMessage() const { return m_errorMessage; }

void AnonSurfBackend::refresh() {
  m_commands.requestStatus();
  m_systemd.requestAnonSurfEnabled();
}

void AnonSurfBackend::start(bool safeKill) {
  beginLifecycle("start", safeKill);
}

void AnonSurfBackend::stop(bool safeKill) { beginLifecycle("stop", safeKill); }

void AnonSurfBackend::restart(bool safeKill) {
  beginLifecycle("restart", safeKill);
}

void AnonSurfBackend::changeIdentity() {
  if (m_busy)
    return;
  invalidateIPCheck();
  beginAction("identity", "Requesting a new Tor identity");
  m_systemd.runUnit("identity", "StartUnit", IdentityUnit);
}

void AnonSurfBackend::checkIP() {
  if (m_busy)
    return;
  beginIPCheck(false);
}

void AnonSurfBackend::beginIPCheck(bool automatic) {
  invalidateIPCheck();
  m_ipCheckStale = false;
  m_autoCheck = automatic;
  beginAction("ip", "Checking public IP and Tor routing");
  m_commands.requestIP();
}

void AnonSurfBackend::setBootEnabled(bool enabled) {
  if (m_busy || enabled == m_bootEnabled)
    return;
  const QString action = enabled ? "boot-enable" : "boot-disable";
  beginAction(action, enabled ? "Enabling AnonSurf at startup"
                              : "Disabling AnonSurf at startup");
  m_systemd.setAnonSurfEnabled(enabled);
}

void AnonSurfBackend::openMonitor() { m_commands.openMonitor(); }

void AnonSurfBackend::dismissError() { setError({}); }
