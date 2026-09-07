#include "anonsurf/anonsurfbackend.h"

namespace {
constexpr auto AnonSurfUnit = "anonsurfd.service";
constexpr int IPCheckDelayMs = 2000;
} // namespace

void AnonSurfBackend::beginLifecycle(const QString &action, bool safeKill) {
  if (m_busy)
    return;
  invalidateIPCheck();
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
  updateProtection();
  if (success) {
    m_progress = successMessage(action);
    if (action == "stop") {
      invalidateIPCheck();
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
  // Give Tor a short bootstrap grace period before the automatic check.
  QTimer::singleShot(IPCheckDelayMs, this, [this]() {
    if (m_busy || !m_statusAvailable || !m_anonsurfActive)
      return;
    beginIPCheck(true);
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
