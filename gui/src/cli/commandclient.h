#pragma once

#include <QObject>
#include <QProcess>
#include <QStringList>

class CommandClient final : public QObject {
  Q_OBJECT

public:
  explicit CommandClient(QObject *parent = nullptr);

  void requestStatus();
  void requestIP();
  void runSafeKill();
  void openMonitor();

signals:
  void statusReady(bool anonsurfActive, bool torActive, const QString &dnsState,
                   const QStringList &dnsAddresses);
  void statusFailed(const QString &message);
  void ipReady(const QString &address, bool isTor);
  void ipFailed(const QString &message);
  void safeKillFinished(bool success, const QString &message);
  void monitorFailed(const QString &message);

private:
  void handleStatusFinished(int exitCode, QProcess::ExitStatus exitStatus);
  void handleIPFinished(int exitCode, QProcess::ExitStatus exitStatus);
  void handleSafeKillFinished(int exitCode, QProcess::ExitStatus exitStatus);
  static QString processError(QProcess &process, const QString &fallback);

  QProcess m_statusProcess;
  QProcess m_ipProcess;
  QProcess m_safeKillProcess;
};
