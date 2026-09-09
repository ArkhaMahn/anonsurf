#pragma once

#include "cli/commandclient.h"
#include "systemd/systemdclient.h"

#include <QObject>
#include <QTimer>

class AnonSurfBackend final : public QObject {
  Q_OBJECT
  Q_PROPERTY(bool statusAvailable READ statusAvailable NOTIFY changed)
  Q_PROPERTY(bool anonsurfActive READ anonsurfActive NOTIFY changed)
  Q_PROPERTY(bool torActive READ torActive NOTIFY changed)
  Q_PROPERTY(bool bootEnabled READ bootEnabled NOTIFY changed)
  Q_PROPERTY(bool busy READ busy NOTIFY changed)
  Q_PROPERTY(QString protectionState READ protectionState NOTIFY changed)
  Q_PROPERTY(QString protectionDetail READ protectionDetail NOTIFY changed)
  Q_PROPERTY(QString dnsMode READ dnsMode NOTIFY changed)
  Q_PROPERTY(QString dnsServers READ dnsServers NOTIFY changed)
  Q_PROPERTY(QString publicIP READ publicIP NOTIFY changed)
  Q_PROPERTY(QString torCheck READ torCheck NOTIFY changed)
  Q_PROPERTY(QString progress READ progress NOTIFY changed)
  Q_PROPERTY(QString errorMessage READ errorMessage NOTIFY changed)

public:
  explicit AnonSurfBackend(QObject *parent = nullptr);

  bool statusAvailable() const;
  bool anonsurfActive() const;
  bool torActive() const;
  bool bootEnabled() const;
  bool busy() const;
  QString protectionState() const;
  QString protectionDetail() const;
  QString dnsMode() const;
  QString dnsServers() const;
  QString publicIP() const;
  QString torCheck() const;
  QString progress() const;
  QString errorMessage() const;

  Q_INVOKABLE void refresh();
  Q_INVOKABLE void start(bool safeKill);
  Q_INVOKABLE void stop(bool safeKill);
  Q_INVOKABLE void restart(bool safeKill);
  Q_INVOKABLE void changeIdentity();
  Q_INVOKABLE void checkIP();
  Q_INVOKABLE void setBootEnabled(bool enabled);
  Q_INVOKABLE void openMonitor();
  Q_INVOKABLE void dismissError();

signals:
  void changed();
  void operationFinished(const QString &action, bool success,
                         const QString &message);

private:
  void beginLifecycle(const QString &action, bool safeKill);
  void runLifecycle(const QString &action);
  void beginAction(const QString &action, const QString &progress);
  void beginIPCheck(bool automatic);
  void invalidateIPCheck();
  void scheduleIPCheck();
  void finishAction(const QString &action, bool success,
                    const QString &message);
  void handleStatus(bool active, bool torActive, const QString &dnsState,
                    const QStringList &addresses);
  void handleAction(const QString &action, bool success,
                    const QString &message);
  void updateProtection();
  void setError(const QString &message);
  static QString lifecycleMethod(const QString &action);
  static QString lifecycleProgress(const QString &action);
  static QString successMessage(const QString &action);

  CommandClient m_commands;
  SystemdClient m_systemd;
  QTimer m_progressTimer;
  bool m_statusAvailable = false;
  bool m_anonsurfActive = false;
  bool m_torActive = false;
  bool m_bootEnabled = false;
  bool m_busy = false;
  bool m_autoCheck = false;
  bool m_ipCheckStale = false;
  QString m_dnsState;
  QStringList m_dnsAddresses;
  QString m_dnsMode = "Unavailable";
  QString m_dnsServers = "No data";
  QString m_protectionState = "UNAVAILABLE";
  QString m_protectionDetail = "Waiting for AnonSurf status.";
  QString m_publicIP = "Not checked";
  QString m_torCheck = "UNKNOWN";
  QString m_progress = "Ready";
  QString m_errorMessage;
  QString m_activeAction;
  QString m_pendingAction;
};
