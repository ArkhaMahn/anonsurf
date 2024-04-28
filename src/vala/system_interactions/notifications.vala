public enum NotifyLevel {
  Ok,
  Warn,
  Error
}

// TODO make it be a class, only call by threads?

public void send_notification(string title, string desc, NotifyLevel level) {
  string icon_name;

  if (level == Ok) {
    icon_name = "security-high";
  } else if (level == Warn) {
    icon_name = "security-medium";
  } else {
    icon_name = "security-low";
  }
  Notify.Notification noti = new Notify.Notification (title, desc, icon_name);
  try {
    noti.show();
  } catch (GLib.Error error) {
    print("Error showing notifications: %s\n", error.message);
  }
}
