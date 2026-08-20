private Subprocess run_systemctl(string action) throws GLib.Error {
  Subprocess process = new Subprocess.newv({
    "/usr/bin/pkexec",
    "/usr/bin/systemctl",
    action,
    "anonsurfd.service"
  }, STDOUT_SILENCE | STDERR_SILENCE);
  process.wait();
  return process;
}

public class ThreadStartAnonSurf {
  public bool is_action_start { private set; get; }

  public ThreadStartAnonSurf(bool is_action_start) {
    this.is_action_start = is_action_start;
  }

  public void run() {
    string action = this.is_action_start ? "start" : "stop";
    string title = this.is_action_start ? "Start AnonSurf" : "Stop AnonSurf";

    try {
      Subprocess process = run_systemctl(action);
      if (process.get_successful()) {
        string message = this.is_action_start ? "AnonSurf is activated" : "AnonSurf is deactivated";
        NotifyLevel level = this.is_action_start ? NotifyLevel.Ok : NotifyLevel.Warn;
        send_notification(title, message, level);
      } else {
        string status = process.get_exit_status().to_string();
        send_notification(title, "Action failed. Exit code: " + status, NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification(title, error.message, NotifyLevel.Error);
    }
  }
}

public class ThreadEnableAnonSurf {
  public bool is_action_enable { private set; get; }

  public ThreadEnableAnonSurf(bool is_action_enable) {
    this.is_action_enable = is_action_enable;
  }

  public void run() {
    string action = this.is_action_enable ? "enable" : "disable";

    try {
      Subprocess process = run_systemctl(action);
      if (process.get_successful()) {
        string message = this.is_action_enable
          ? "AnonSurf is enabled at boot"
          : "AnonSurf is no longer enabled at boot";
        NotifyLevel level = this.is_action_enable ? NotifyLevel.Ok : NotifyLevel.Warn;
        send_notification("AnonSurf startup", message, level);
      } else {
        string status = process.get_exit_status().to_string();
        send_notification("AnonSurf startup", "Action failed. Exit code: " + status, NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("AnonSurf startup", error.message, NotifyLevel.Error);
    }
  }
}

public class ThreadMyIP {
  public void run() {
    try {
      Subprocess process = new Subprocess.newv({
        "/usr/bin/anonsurf",
        "myip"
      }, STDOUT_PIPE | STDERR_PIPE);
      string? stdout_buf;
      string? stderr_buf;
      process.communicate_utf8(null, null, out stdout_buf, out stderr_buf);
      string stdout_text = stdout_buf ?? "";
      string stderr_text = stderr_buf ?? "Unknown error";

      if (!process.get_successful()) {
        send_notification("IP check", stderr_text.strip(), NotifyLevel.Error);
      } else if (stdout_text.contains("Tor:       yes")) {
        send_notification("IP check", "Your connection is under Tor network", NotifyLevel.Ok);
      } else if (stdout_text.contains("Tor:       no")) {
        send_notification("IP check", "Your connection is not under Tor network", NotifyLevel.Warn);
      } else {
        send_notification("IP check", "Unexpected AnonSurf response", NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("IP check", error.message, NotifyLevel.Error);
    }
  }
}
