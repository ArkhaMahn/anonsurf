#include "system/systemdclient.h"

void SystemdClient::handleJobRemoved(uint, const QDBusObjectPath &job,
                                     const QString &unit,
                                     const QString &result) {
  const QString jobPath = job.path();
  if (m_pendingJobs.contains(jobPath)) {
    finishJob(m_pendingJobs.take(jobPath), result);
    return;
  }
  if (m_pendingRequests.contains(unit))
    m_completedJobs.insert(jobPath, {unit, result});
}

void SystemdClient::finishJob(const QString &action, const QString &result) {
  const bool success = result == "done";
  emit actionFinished(action, success,
                      success ? QString{} : jobError(action, result));
}

void SystemdClient::clearCompletedJobs(const QString &unit) {
  for (auto job = m_completedJobs.begin(); job != m_completedJobs.end();) {
    if (job->unit == unit)
      job = m_completedJobs.erase(job);
    else
      ++job;
  }
}

QString SystemdClient::jobError(const QString &action, const QString &result) {
  if (action == "identity") {
    return QString("Tor could not change identity (systemd result: %1).")
        .arg(result);
  }
  return QString("AnonSurf could not complete the %1 operation "
                 "(systemd result: %2).")
      .arg(action, result);
}
