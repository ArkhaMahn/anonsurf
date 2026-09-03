#include "backend/anonsurfbackend.h"

namespace {
constexpr auto AnonSurfUnit = "anonsurfd.service";
constexpr int IPCheckDelayMs = 2000;
} // namespace

void AnonSurfBackend::beginLifecycle(const QString &action, bool safeKill) {
  if (m_busy)
    return;
  m_publicIP = "Not checked";
  m_torCheck = "UNKNOWN";
  if (safeKill) {
    m_pendingAction = action;
    beginAction("safekill",
                "Closing networked applications and clearing caches");
    m_commands.runSafeKill();
    return;
  }
  runLifecycle(action);
}

void AnonSurfBackend::runLifecycle(const QString &action) {
  beginAction(action, lifecycleProgress(action));
  m_progressTimer.start();
  m_systemd.runUnit(action, lifecycleMethod(action), AnonSurfUnit);
}

void AnonSurfBackend::beginAction(const QString &action,
                                  const QString &progress) {
  m_busy = true;
  m_activeAction = action;
  m_progress = progress;
  setError({});
  emit changed();
}

void AnonSurfBackend::finishAction(const QString &action, bool success,
                                   const QString &message) {
  const bool automatic = action == "ip" && m_autoCheck;
  m_autoCheck = false;
  m_busy = false;
  m_activeAction.clear();
  m_progressTimer.stop();
  if (success) {
    m_progress = successMessage(action);
    if (action == "stop") {
      m_publicIP = "Not checked";
      m_torCheck = "UNKNOWN";
    }
    if (action == "start" || action == "restart" || action == "identity")
      scheduleIPCheck();
  } else if (automatic) {
    // An unattended check must not raise a modal error the user did not ask
    // for; the unknown result is already visible in the overview.
    m_progress = "Public IP check unavailable";
  } else {
    m_progress = "Operation failed";
    setError(message);
  }
  emit changed();
  refresh();
}

void AnonSurfBackend::scheduleIPCheck() {
  // Routing needs a moment to settle after the unit reports success, otherwise
  // the request can still leave outside Tor and report a misleading address.
  QTimer::singleShot(IPCheckDelayMs, this, [this]() {
    if (m_busy || !m_anonsurfActive)
      return;
    m_autoCheck = true;
    beginAction("ip", "Checking public IP and Tor routing");
    m_commands.requestIP();
  });
}

void AnonSurfBackend::handleAction(const QString &action, bool success,
                                   const QString &message) {
  if (action == "safekill" && success && !m_pendingAction.isEmpty()) {
    const QString nextAction = m_pendingAction;
    m_pendingAction.clear();
    runLifecycle(nextAction);
    return;
  }
  m_pendingAction.clear();
  finishAction(action, success, message);
}

void AnonSurfBackend::setError(const QString &message) {
  if (m_errorMessage == message)
    return;
  m_errorMessage = message;
  emit changed();
}
