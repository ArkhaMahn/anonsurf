#include "anonsurf/anonsurfbackend.h"
#include "application/instance.h"
#include "tray/traycontroller.h"

#include <QApplication>
#include <QDebug>
#include <QIcon>
#include <QMessageBox>
#include <QQmlApplicationEngine>
#include <QQuickStyle>
#include <QWindow>

#ifndef ANONSURF_VERSION
#define ANONSURF_VERSION "6.0.0"
#endif

int main(int argc, char *argv[]) {
  QQuickStyle::setStyle("Fusion");

  QApplication app(argc, argv);
  app.setApplicationName("anonsurf-gui");
  app.setApplicationDisplayName("AnonSurf");
  app.setOrganizationName("ParrotSec");
  app.setApplicationVersion(ANONSURF_VERSION);
  app.setDesktopFileName("org.parrotsec.Anonsurf");
  app.setWindowIcon(
      QIcon::fromTheme("anonsurf", QIcon(":/icons/anonsurf.png")));

  DesktopInstance instance;
  const auto startup = instance.start();
  if (startup == DesktopInstance::Startup::Activated)
    return 0;

  AnonSurfBackend backend;
  TrayController desktop(backend);
  QQmlApplicationEngine engine;
  engine.setInitialProperties(
      {{"backend", QVariant::fromValue(static_cast<QObject *>(&backend))},
       {"desktop", QVariant::fromValue(static_cast<QObject *>(&desktop))}});
  engine.load(QUrl("qrc:/qt/qml/Parrot/AnonSurf/qml/app/Main.qml"));
  if (engine.rootObjects().isEmpty()) {
    qCritical() << "Failed to load the AnonSurf interface with Qt" << qVersion();
    QMessageBox::critical(
        nullptr, QStringLiteral("AnonSurf"),
        QStringLiteral("The AnonSurf interface could not be loaded.\n\n"
                       "Run \"anonsurf-gui\" in a terminal to see the details."));
    return 1;
  }

  auto *window = qobject_cast<QWindow *>(engine.rootObjects().first());
  if (!window) {
    qCritical() << "The main AnonSurf window is unavailable.";
    QMessageBox::critical(
        nullptr, QStringLiteral("AnonSurf"),
        QStringLiteral("The main AnonSurf window could not be created.\n\n"
                       "Run \"anonsurf-gui\" in a terminal to see the details."));
    return 1;
  }
  desktop.attachWindow(window);
  QObject::connect(&instance, &DesktopInstance::activationRequested, &desktop,
                   &TrayController::showWindow);

  return app.exec();
}
