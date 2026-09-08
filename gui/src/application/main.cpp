#include "anonsurf/anonsurfbackend.h"
#include "application/instance.h"

#include <QDebug>
#include <QGuiApplication>
#include <QIcon>
#include <QQmlApplicationEngine>
#include <QQuickStyle>
#include <QWindow>

#ifndef ANONSURF_VERSION
#define ANONSURF_VERSION "6.0.0"
#endif

int main(int argc, char *argv[]) {
  QQuickStyle::setStyle("Fusion");

  QGuiApplication app(argc, argv);
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
  QQmlApplicationEngine engine;
  engine.setInitialProperties(
      {{"backend", QVariant::fromValue(static_cast<QObject *>(&backend))}});
  engine.load(QUrl("qrc:/qt/qml/Parrot/AnonSurf/qml/app/Main.qml"));
  if (engine.rootObjects().isEmpty())
    return 1;

  auto *window = qobject_cast<QWindow *>(engine.rootObjects().first());
  if (!window) {
    qCritical() << "The main AnonSurf window is unavailable.";
    return 1;
  }
  QObject::connect(&instance, &DesktopInstance::activationRequested, window,
                   [window, &backend]() {
                     window->showNormal();
                     window->raise();
                     window->requestActivate();
                     backend.refresh();
                   });

  return app.exec();
}
