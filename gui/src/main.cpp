#include "backend/anonsurfbackend.h"

#include <QGuiApplication>
#include <QIcon>
#include <QQmlApplicationEngine>
#include <QQuickStyle>

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

  AnonSurfBackend backend;
  QQmlApplicationEngine engine;
  engine.setInitialProperties(
      {{"backend", QVariant::fromValue(static_cast<QObject *>(&backend))}});
  engine.load(QUrl("qrc:/qt/qml/Parrot/AnonSurf/qml/app/Main.qml"));
  if (engine.rootObjects().isEmpty())
    return 1;

  return app.exec();
}
