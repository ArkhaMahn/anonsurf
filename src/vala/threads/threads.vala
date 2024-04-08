
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
        var action_start_anonsurf = new Subprocess.newv({"service", "anonsurfd", "start"}, NONE);
        action_start_anonsurf.wait();
      } else {
        var action_start_anonsurf = new Subprocess.newv({"service", "anonsurfd", "stop"}, NONE);
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

    public void run() {
      if (this.is_action_enable == true) {
        var action_start_anonsurf = new Subprocess.newv({"systemctl", "enable", "anonsurfd"}, NONE);
        action_start_anonsurf.wait();
      } else {
        var action_start_anonsurf = new Subprocess.newv({"systemctl", "disable", "anonsurfd"}, NONE);
        action_start_anonsurf.wait();
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
