
public class WorkerSystemdService {
    public bool is_action_start { private set; get; }

    public WorkerSystemdService(bool is_action_start) {
      if (is_action_start) {
        this.is_action_start = true;
      } else {
        this.is_action_start = false;
      }
    }

    public void run() {
      if (this.is_action_start == true) {
        Subprocess action_start_anonsurf = new Subprocess.newv({"service", "anonsurfd", "start"}, NONE);
        action_start_anonsurf.wait();
      } else {
        Subprocess action_start_anonsurf = new Subprocess.newv({"service", "anonsurfd", "stop"}, NONE);
        action_start_anonsurf.wait();
      }
    }
}


public class WorkerSystemdStartup {
      public bool is_action_enable { private set; get; }

    public WorkerSystemdStartup(bool is_action_enable) {
      if (is_action_enable) {
        this.is_action_enable = true;
      } else {
        this.is_action_enable = false;
      }
    }

    private void enable_anonsurf() {
      // TODO improve output messages
      Subprocess action_enable_anonsurf = new Subprocess.newv({"systemctl", "enable", "anonsurfd"}, NONE);
      action_enable_anonsurf.wait();

      int exit_code = action_enable_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("Enabled", "AnonSurf is enabled at boot", NotifyLevel.Ok);
      } else {
        send_notification("Failed to enable", "AnonSurf couldn't be enabled at boot", NotifyLevel.Error);
      }
    }

    private void disable_anonsurf() {
      Subprocess action_disable_anonsurf = new Subprocess.newv({"systemctl", "disable", "anonsurfd"}, NONE);
      action_disable_anonsurf.wait();

      int exit_code = action_disable_anonsurf.get_exit_status();
      if (exit_code == 0) {
        send_notification("Disabled", "AnonSurf is no longer enabled at boot", NotifyLevel.Warn);
      } else {
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


public class WorkerCheckTor {
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
