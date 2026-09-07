#include "backend/anonsurfbackend.h"

void AnonSurfBackend::handleStatus(bool active, bool torActive,
                                   const QString &dnsState,
                                   const QStringList &addresses) {
  const bool routingChanged =
      m_statusAvailable && (active != m_anonsurfActive ||
                            torActive != m_torActive || dnsState != m_dnsState ||
                            addresses != m_dnsAddresses);
  if (routingChanged)
    invalidateIPCheck();
  m_statusAvailable = true;
  m_anonsurfActive = active;
  m_torActive = torActive;
  m_dnsState = dnsState;
  m_dnsAddresses = addresses;
  m_dnsServers = addresses.isEmpty() ? "No nameservers" : addresses.join(", ");
  if (dnsState == "tor")
    m_dnsMode = "Tor DNS";
  else if (dnsState == "tor-redirect")
    m_dnsMode = "Tor redirection";
  else if (dnsState == "clearnet")
    m_dnsMode = "Clearnet";
  else if (dnsState == "localhost-only")
    m_dnsMode = "Localhost only";
  else if (dnsState == "missing")
    m_dnsMode = "Missing resolver";
  else if (dnsState == "empty")
    m_dnsMode = "Empty resolver";
  else
    m_dnsMode = "Unknown";
  updateProtection();
  emit changed();
}

void AnonSurfBackend::updateProtection() {
  if (!m_statusAvailable) {
    m_protectionState = "UNAVAILABLE";
    return;
  }
  const bool torDNSConfigured =
      m_dnsState == "tor" || m_dnsState == "tor-redirect";
  const bool brokenDNS = m_dnsState == "localhost-only" ||
                         m_dnsState == "missing" || m_dnsState == "empty";
  if (m_anonsurfActive && m_torCheck == "CLEARNET") {
    m_protectionState = "DEGRADED";
    m_protectionDetail =
        "AnonSurf is active, but the public IP check reported clearnet.";
  } else if (m_anonsurfActive && m_torActive && torDNSConfigured) {
    if (m_torCheck == "TOR") {
      m_protectionState = "ROUTED";
      m_protectionDetail = "The last public IP check reported a Tor exit.";
    } else {
      m_protectionState = "UNVERIFIED";
      m_protectionDetail =
          "Services are active; the public IP route is not confirmed.";
    }
  } else if (m_anonsurfActive || brokenDNS) {
    m_protectionState = "DEGRADED";
    m_protectionDetail =
        m_anonsurfActive
            ? "AnonSurf is active, but Tor or DNS needs attention."
            : "AnonSurf is inactive and DNS needs attention.";
  } else {
    m_protectionState = "INACTIVE";
    m_protectionDetail = "AnonSurf service is inactive.";
  }
}

void AnonSurfBackend::invalidateIPCheck() {
  m_ipCheckStale = true;
  m_publicIP = "Not checked";
  m_torCheck = "UNKNOWN";
  updateProtection();
}

QString AnonSurfBackend::lifecycleMethod(const QString &action) {
  if (action == "stop")
    return "StopUnit";
  if (action == "restart")
    return "ReloadOrRestartUnit";
  return "StartUnit";
}

QString AnonSurfBackend::lifecycleProgress(const QString &action) {
  if (action == "stop")
    return "Restoring the original network";
  if (action == "restart")
    return "Restarting protected routing";
  return "Starting AnonSurf";
}

QString AnonSurfBackend::successMessage(const QString &action) {
  if (action == "start")
    return "AnonSurf started";
  if (action == "stop")
    return "Original network restored";
  if (action == "restart")
    return "AnonSurf restarted";
  if (action == "identity")
    return "New Tor identity requested";
  if (action == "ip")
    return "Tor check completed";
  if (action == "boot-enable")
    return "AnonSurf enabled at startup";
  if (action == "boot-disable")
    return "AnonSurf disabled at startup";
  return "Operation completed";
}
