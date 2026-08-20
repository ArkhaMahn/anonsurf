/*
Create thread pools that start, stop AnonSurf; enable / disable AnonSurf at boot or
Check current Public IP and see if it's under Tor network
*/


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
      Subprocess action_start_anonsurf = new Subprocess.newv({"/usr/sbin/service", "anonsurfd", "start"}, STDOUT_SILENCE | STDERR_SILENCE);
      action_start_anonsurf.wait();

      int exit_code = action_start_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("Start AnonSurf", "AnonSurf is activated", NotifyLevel.Ok);
      } else {
        send_notification("Start AnonSurf", "AnonSurf couldn't start. Exit code: " + exit_code.to_string(), NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("Start AnonSurf", error.message, NotifyLevel.Error);
    }
  }


  private void stop_anonsurf() {
    try {
      Subprocess action_stop_anonsurf = new Subprocess.newv({"/usr/sbin/service", "anonsurfd", "stop"}, STDOUT_SILENCE | STDERR_SILENCE);
      action_stop_anonsurf.wait();

      int exit_code = action_stop_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("Stop AnonSurf", "AnonSurf is deactivated", NotifyLevel.Warn);
      } else {
        send_notification("Stop AnonSurf", "AnonSurf couldn't stop. Exit code: "  + exit_code.to_string(), NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("Stop AnonSurf", error.message, NotifyLevel.Error);
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
      Subprocess action_enable_anonsurf = new Subprocess.newv({"/usr/bin/systemctl", "enable", "anonsurfd"}, STDOUT_SILENCE | STDERR_SILENCE);
      action_enable_anonsurf.wait();

      int exit_code = action_enable_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("AnonSurf startup", "AnonSurf is enabled at boot", NotifyLevel.Ok);
      } else {
        send_notification("AnonSurf startup", "AnonSurf failed to disable at boot. Exit code: " + exit_code.to_string(), NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("AnonSurf startup", error.message, NotifyLevel.Error);
    }
  }

  private void disable_anonsurf() {
    try {
      Subprocess action_disable_anonsurf = new Subprocess.newv({"/usr/bin/systemctl", "disable", "anonsurfd"}, STDOUT_SILENCE | STDERR_SILENCE);
      action_disable_anonsurf.wait();

      int exit_code = action_disable_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("AnonSurf startup", "AnonSurf is no longer enabled at boot", NotifyLevel.Warn);
      } else {
        send_notification("AnonSurf startup", "AnonSurf failed to disable at boot. Exit code: " + exit_code.to_string(), NotifyLevel.Error);
      }
    } catch (GLib.Error error) {
      send_notification("AnonSurf startup", error.message, NotifyLevel.Error);
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
    Soup.Session session = new Soup.Session();
    Soup.Message message = new Soup.Message("GET", "https://check.torproject.org/");

    session.send_message(message);
    string recv_data = (string)message.response_body.data;

    if (recv_data.contains("Congratulations. This browser is configured to use Tor.")) {
      send_notification("IP check", "Your connection is under Tor network", NotifyLevel.Ok);
    }
    else if (recv_data.contains("Sorry. You are not using Tor.")) {
      send_notification("IP check", "Your connection is not under Tor network", NotifyLevel.Warn);
    } else {
      send_notification("CIP check", "Unexpected response data", NotifyLevel.Error);
    }
  }
}
