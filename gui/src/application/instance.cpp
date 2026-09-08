#include "application/instance.h"

#include <QDBusConnectionInterface>
#include <QDBusMessage>
#include <QDBusReply>
#include <QDebug>

namespace {
constexpr auto Service = "org.parrotsec.Anonsurf";
constexpr auto Path = "/org/parrotsec/Anonsurf";
} // namespace

DesktopInstance::DesktopInstance(const QDBusConnection &bus, QObject *parent)
    : QObject(parent), m_bus(bus) {}

DesktopInstance::~DesktopInstance() {
  if (m_registered)
    m_bus.unregisterService(Service);
  if (m_exported)
    m_bus.unregisterObject(Path);
}

DesktopInstance::Startup DesktopInstance::start() {
  if (!m_bus.isConnected()) {
    qWarning() << "The session bus is unavailable; single-instance activation "
                  "is disabled.";
    return Startup::Unavailable;
  }
  m_exported =
      m_bus.registerObject(Path, this, QDBusConnection::ExportScriptableSlots);
  if (!m_exported) {
    qCritical() << "Cannot export the AnonSurf activation endpoint:"
                << m_bus.lastError().message();
    return Startup::Failed;
  }

  const auto reply = m_bus.interface()->registerService(
      Service, QDBusConnectionInterface::DontQueueService);
  if (!reply.isValid()) {
    qCritical() << "Cannot register the AnonSurf desktop instance:"
                << reply.error().message();
    return Startup::Failed;
  }
  if (reply.value() == QDBusConnectionInterface::ServiceRegistered) {
    m_registered = true;
    return Startup::Primary;
  }

  const auto activation = m_bus.call(
      QDBusMessage::createMethodCall(Service, Path, Service, "Activate"),
      QDBus::Block, 3000);
  if (activation.type() == QDBusMessage::ErrorMessage) {
    qCritical() << "Cannot activate the existing AnonSurf window:"
                << activation.errorMessage();
    return Startup::Failed;
  }
  return Startup::Activated;
}

void DesktopInstance::Activate() { emit activationRequested(); }
