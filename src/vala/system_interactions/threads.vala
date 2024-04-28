

public class ThreadStartAnonSurf {
  public bool is_action_start { private set; get; }

  public ThreadStartAnonSurf(bool is_action_start) {
    if (is_action_start) {
      this.is_action_start = true;
    } else {
      this.is_action_start = false;
    }
  }

  private void start_anonsurf() {
    try {
      Subprocess action_start_anonsurf = new Subprocess.newv({"service", "anonsurfd", "start"}, NONE);
      action_start_anonsurf.wait();

      int exit_code = action_start_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("Activated", "AnonSurf is activated", NotifyLevel.Ok);
      } else {
        send_notification("Failed to activated", "AnonSurf couldn't start", NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("Failed to activated", "AnonSurf couldn't start", NotifyLevel.Error);
    }
  }


  private void stop_anonsurf() {
    try {
      Subprocess action_stop_anonsurf = new Subprocess.newv({"service", "anonsurfd", "stop"}, NONE);
      action_stop_anonsurf.wait();

      int exit_code = action_stop_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("Deactivated", "AnonSurf is deactivated", NotifyLevel.Warn);
      } else {
        send_notification("Failed to deactivate", "AnonSurf couldn't stop", NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("Failed to deactivate", "AnonSurf couldn't stop", NotifyLevel.Error);
    }
  }

  public void run() {
    if (this.is_action_start == true) {
      start_anonsurf();
    } else {
      stop_anonsurf();
    }
  }
}


public class ThreadEnableAnonSurf {
  // TODO improve output messages
  public bool is_action_enable { private set; get; }

  public ThreadEnableAnonSurf(bool is_action_enable) {
    if (is_action_enable) {
      this.is_action_enable = true;
    } else {
      this.is_action_enable = false;
    }
  }

  /*
    Given flags STDOUT_SILENCE | STDERR_SILENCE -> No StdOut nor StdErr shows
  */

  private void enable_anonsurf() {
    try {
      Subprocess action_enable_anonsurf = new Subprocess.newv({"systemctl", "enable", "anonsurfd"}, STDOUT_SILENCE | STDERR_SILENCE);
      action_enable_anonsurf.wait();

      int exit_code = action_enable_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("Enabled", "AnonSurf is enabled at boot", NotifyLevel.Ok);
      } else {
        send_notification("Failed to Disable", "AnonSurf failed to disable at boot", NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("Failed to enable", "AnonSurf couldn't be enabled at boot", NotifyLevel.Error);
    }
  }

  private void disable_anonsurf() {
    try {
      Subprocess action_disable_anonsurf = new Subprocess.newv({"systemctl", "disable", "anonsurfd"}, STDOUT_SILENCE | STDERR_SILENCE);
      action_disable_anonsurf.wait();

      int exit_code = action_disable_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("Disabled", "AnonSurf is no longer enabled at boot", NotifyLevel.Warn);
      } else {
        send_notification("Failed to Disable", "AnonSurf failed to disable at boot", NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("Failed to Disable", "AnonSurf failed to disable at boot", NotifyLevel.Error);
    }
  }

  public void run() {
    if (this.is_action_enable == true) {
      enable_anonsurf();
    } else {
      disable_anonsurf();
    }
  }
}


public class ThreadMyIP {
  public void run() {
    /*
      https://wiki.gnome.org/Projects/Vala/LibSoupSample
    */
    var session = new Soup.Session();
    var message = new Soup.Message("GET", "https://check.torproject.org/");

    session.send_message(message);

    if (((string)message.response_body.data).contains("Congratulations. This browser is configured to use Tor.")) {
      send_notification("Connection is secured", "Your connection is under Tor network", NotifyLevel.Ok);
    }
    else if (((string)message.response_body.data).contains("Sorry. You are not using Tor.")) {
      send_notification("Connection is not secured", "Your connection is not under Tor network", NotifyLevel.Warn);
    } else {
      send_notification("Check Error", "Unexpected error while checking network connection", NotifyLevel.Error);
    }
  }
}
