/*
  Dbus interface for anonsurfd service
  Create DBus interface: https://wiki.gnome.org/Projects/Vala/DBusClientSamples
  Working with Unit file? https://pi3g.com/enabling-and-disabling-a-systemd-service-in-python-using-dbus/
  Example using DBus https://stackoverflow.com/questions/43499880/how-to-extract-service-state-via-systemd-dbus-api
  Example using dbus https://stackoverflow.com/questions/32332128/how-to-use-polkit-with-python-to-run-systemd-d-bus-commands-from-non-root-user. Python code actually worked
  Output example: dbus.Struct((dbus.Boolean(True), dbus.Boolean(False), dbus.Dictionary({dbus.String('polkit.retains_authorization_after_challenge'): dbus.String('true'), dbus.String('polkit.temporary_authorization_id'): dbus.String('tmpauthz1')}, signature=dbus.Signature('ss'))), signature=None)
there are pending jobs, lets wait for them to finish.
inactive dead
  The script also has a demotration of stopping a service
  To get status of a service, simply run active_state = prop_unit.Get('org.freedesktop.systemd1.Unit', 'ActiveState'). The value is activate
  Is is possible to disable boot with DisableUnitFiles?
*/

// name of Interface in dbus_path, in this case, it is org.freedesktops.systemd1.Unit
// [DBus (name = "org.freedesktops.systemd1.Unit")]

// interface AnonsurfIface: Object {
//   // Methods in Interface
//   public signal void stop(string mode);
//   // Restart is needed?
//   // Check status is better than file exists?
// }


// public class AnsurfBus() {
//   private string dbus_name = "org.freeesktop.systemd1";
//   private string dbus_path = "/org/freedesktop/systemd1/unit/anonsurfd_2eservice";

// }

SYSTEMD_BUSNAME = "org.freedesktop.systemd1"
SYSTEMD_PATH = "/org/freedesktop/systemd1"
SYSTEMD_MANAGER_INTERFACE = "org.freedesktop.systemd1.Manager"
SYSTEMD_UNIT_INTERFACE = "org.freedesktop.systemd1.Unit"
