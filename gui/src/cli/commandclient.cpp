#include "cli/commandclient.h"

#include <QDebug>
#include <QFileInfo>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>

namespace {
constexpr auto AnonSurfPath = "/usr/bin/anonsurf";
constexpr auto SafeKillPath = "/usr/lib/anonsurf/safekill";
constexpr auto KonsolePath = "/usr/bin/konsole";
} // namespace

CommandClient::CommandClient(QObject *parent) : QObject(parent) {
  connect(&m_statusProcess, &QProcess::finished, this,
          &CommandClient::handleStatusFinished);
  connect(&m_ipProcess, &QProcess::finished, this,
          &CommandClient::handleIPFinished);
  connect(&m_safeKillProcess, &QProcess::finished, this,
          &CommandClient::handleSafeKillFinished);

  connect(&m_statusProcess, &QProcess::errorOccurred, this,
          [this](QProcess::ProcessError error) {
            if (error == QProcess::FailedToStart)
              emit statusFailed("The AnonSurf CLI could not be started.");
          });
  connect(&m_ipProcess, &QProcess::errorOccurred, this,
          [this](QProcess::ProcessError error) {
            if (error == QProcess::FailedToStart)
              emit ipFailed("The AnonSurf CLI could not be started.");
          });
  connect(&m_safeKillProcess, &QProcess::errorOccurred, this,
          [this](QProcess::ProcessError error) {
            if (error == QProcess::FailedToStart) {
              emit safeKillFinished(false, "SafeKill could not be started.");
            }
          });
}

CommandClient::~CommandClient() {
  // QProcess can emit completion signals while its owner is being destroyed.
  for (auto *process : {&m_statusProcess, &m_ipProcess, &m_safeKillProcess}) {
    process->disconnect(this);
    if (process->state() != QProcess::NotRunning) {
      process->kill();
      if (!process->waitForFinished(1000))
        qWarning() << "Cannot finish GUI command shutdown:"
                   << process->errorString();
    }
  }
}

void CommandClient::requestStatus() {
  if (m_statusProcess.state() != QProcess::NotRunning)
    return;
  m_statusProcess.start(AnonSurfPath, {"status", "--json"});
}

void CommandClient::requestIP() {
  if (m_ipProcess.state() != QProcess::NotRunning)
    return;
  m_ipProcess.start(AnonSurfPath, {"ip", "--json"});
}

void CommandClient::runSafeKill() {
  if (m_safeKillProcess.state() != QProcess::NotRunning)
    return;
  m_safeKillProcess.start(SafeKillPath);
}

void CommandClient::openMonitor() {
  if (!QFileInfo(KonsolePath).isExecutable()) {
    emit monitorFailed("Konsole is required to open the Tor monitor.");
    return;
  }
  if (!QProcess::startDetached(KonsolePath, {"-e", AnonSurfPath, "monitor"})) {
    emit monitorFailed("Konsole could not start the Tor monitor.");
  }
}

void CommandClient::handleStatusFinished(int exitCode,
                                         QProcess::ExitStatus exitStatus) {
  if (exitStatus != QProcess::NormalExit || exitCode != 0) {
    emit statusFailed(
        processError(m_statusProcess, "AnonSurf status is unavailable."));
    return;
  }

  QJsonParseError parseError;
  const QJsonDocument document = QJsonDocument::fromJson(
      m_statusProcess.readAllStandardOutput(), &parseError);
  if (parseError.error != QJsonParseError::NoError || !document.isObject()) {
    emit statusFailed("AnonSurf returned an invalid status response.");
    return;
  }

  const QJsonObject object = document.object();
  if (!object.value("anonsurf_active").isBool() ||
      !object.value("tor_active").isBool() ||
      !object.value("dns_state").isString() ||
      !object.value("dns_addresses").isArray()) {
    emit statusFailed("AnonSurf returned an incomplete status response.");
    return;
  }

  QStringList addresses;
  for (const QJsonValue &value : object.value("dns_addresses").toArray()) {
    if (value.isString())
      addresses.append(value.toString());
  }
  emit statusReady(object.value("anonsurf_active").toBool(),
                   object.value("tor_active").toBool(),
                   object.value("dns_state").toString(), addresses);
}

void CommandClient::handleIPFinished(int exitCode,
                                     QProcess::ExitStatus exitStatus) {
  if (exitStatus != QProcess::NormalExit || exitCode != 0) {
    emit ipFailed(processError(m_ipProcess, "The Tor check failed."));
    return;
  }

  QJsonParseError parseError;
  const QJsonDocument document =
      QJsonDocument::fromJson(m_ipProcess.readAllStandardOutput(), &parseError);
  const QJsonObject object = document.object();
  if (parseError.error != QJsonParseError::NoError || !document.isObject() ||
      !object.value("ip").isString() || !object.value("is_tor").isBool()) {
    emit ipFailed("AnonSurf returned an invalid Tor check response.");
    return;
  }
  emit ipReady(object.value("ip").toString(), object.value("is_tor").toBool());
}

void CommandClient::handleSafeKillFinished(int exitCode,
                                           QProcess::ExitStatus exitStatus) {
  const bool success = exitStatus == QProcess::NormalExit && exitCode == 0;
  emit safeKillFinished(
      success,
      success
          ? QString{}
          : processError(m_safeKillProcess,
                         "SafeKill failed. Ensure BleachBit is installed."));
}

QString CommandClient::processError(QProcess &process,
                                    const QString &fallback) {
  const QString stderrText =
      QString::fromUtf8(process.readAllStandardError()).trimmed();
  if (stderrText.isEmpty())
    return fallback;

  const QStringList lines = stderrText.split('\n', Qt::SkipEmptyParts);
  for (auto line = lines.crbegin(); line != lines.crend(); ++line) {
    if (line->startsWith("[FAIL] "))
      return line->mid(7).trimmed();
  }
  return stderrText;
}
