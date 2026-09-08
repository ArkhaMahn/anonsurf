#include "tray/traymenu.h"

#include <QAction>

TrayMenu::TrayMenu(QWidget *parent) : QMenu(parent) {
  setToolTipsVisible(true);
  m_status = addAction("AnonSurf: UNAVAILABLE");
  m_status->setEnabled(false);
  addAction("Show AnonSurf", this, &TrayMenu::showRequested);
  addSeparator();
  m_toggle = addAction("Start AnonSurf", this, [this]() {
    emit lifecycleRequested(m_active ? "stop" : "start");
  });
  m_restart = addAction("Restart", this,
                        [this]() { emit lifecycleRequested("restart"); });
  m_identity = addAction("New Identity", this, &TrayMenu::identityRequested);
  m_ip = addAction("Verify IP", this, &TrayMenu::ipRequested);
  m_monitor = addAction("Tor Monitor", this, &TrayMenu::monitorRequested);
  addSeparator();
  m_quit = addAction("Quit", this, &TrayMenu::quitRequested);
  m_quit->setToolTip("Quit the interface without stopping AnonSurf routing.");
  m_toggle->setToolTip("Uses the SafeKill option selected in the main window.");
  m_restart->setToolTip(
      "Uses the SafeKill option selected in the main window.");
  m_identity->setToolTip(
      "Request new circuits; a different exit IP is not guaranteed.");
  updateState({});
}

void TrayMenu::updateState(const TrayState &state) {
  m_active = state.active;
  m_status->setText("AnonSurf: " + state.label);
  m_status->setToolTip(state.detail);
  m_toggle->setText(state.active ? "Stop AnonSurf" : "Start AnonSurf");
  const bool ready = state.available && !state.busy;
  m_toggle->setEnabled(ready);
  m_restart->setEnabled(ready && state.active);
  m_identity->setEnabled(ready && state.active && state.torActive);
  m_ip->setEnabled(ready);
  m_monitor->setEnabled(ready && state.active && state.torActive);
  m_quit->setEnabled(!state.busy);
}
