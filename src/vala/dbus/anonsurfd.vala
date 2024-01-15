/*
  Dbus interface for anonsurfd service
  Create DBus interface: https://wiki.gnome.org/Projects/Vala/DBusClientSamples
  Working with Unit file? https://pi3g.com/enabling-and-disabling-a-systemd-service-in-python-using-dbus/
  Example using DBus https://stackoverflow.com/questions/43499880/how-to-extract-service-state-via-systemd-dbus-api
*/

// name of Interface in dbus_path, in this case, it is org.freedesktops.systemd1.Unit
[DBus (name = "org.freedesktops.systemd1.Unit")]

interface AnonsurfIface: Object {
  // Methods in Interface
  public signal void stop(string mode);
  // Restart is needed?
  // Check status is better than file exists?
}


public class AnsurfBus() {
  private string dbus_name = "org.freeesktop.systemd1";
  private string dbus_path = "/org/freedesktop/systemd1/unit/anonsurfd_2eservice";

}