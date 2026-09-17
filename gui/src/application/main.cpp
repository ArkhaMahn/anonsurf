#include "anonsurf/anonsurfbackend.h"
#include "application/instance.h"
#include "tray/traycontroller.h"

#include <QApplication>
#include <QDebug>
#include <QIcon>
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
  if (startup == DesktopInstance::Startup::Failed)
    return 1;

  AnonSurfBackend backend;
  TrayController desktop(backend);
  QQmlApplicationEngine engine;
  engine.setInitialProperties(
      {{"backend", QVariant::fromValue(static_cast<QObject *>(&backend))},
       {"desktop", QVariant::fromValue(static_cast<QObject *>(&desktop))}});
  engine.load(QUrl("qrc:/qt/qml/Parrot/AnonSurf/qml/app/Main.qml"));
  if (engine.rootObjects().isEmpty())
    return 1;

  auto *window = qobject_cast<QWindow *>(engine.rootObjects().first());
  if (!window) {
    qCritical() << "The main AnonSurf window is unavailable.";
    return 1;
  }
  desktop.attachWindow(window);
  QObject::connect(&instance, &DesktopInstance::activationRequested, &desktop,
                   &TrayController::showWindow);

  return app.exec();
}
