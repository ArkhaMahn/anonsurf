using Gtk;

public class ButtonChangeID: Button {
  public ButtonChangeID() {
    this.set_label("Change ID");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
  }

  public async void on_click_button() {
    try {
      Subprocess process = new Subprocess.newv({
        "/usr/bin/pkexec",
        "/usr/lib/anonsurf/anondaemon",
        "changeid"
      }, STDOUT_SILENCE | STDERR_SILENCE);
      yield process.wait_async();

      if (process.get_successful()) {
        send_notification("Change ID succeeded", "ID successfully changed", NotifyLevel.Ok);
      } else {
        string status = process.get_exit_status().to_string();
        send_notification("Change ID failed", "Exit code: " + status, NotifyLevel.Error);
      }
    } catch (Error error) {
      send_notification("Change ID failed", error.message, NotifyLevel.Error);
    }
  }

  public void set_surf_activated() {
    this.set_sensitive(true);
  }

  public void set_surf_deactivated() {
    this.set_sensitive(false);
  }
}

public class ButtonTorController: Button {
  public DialogTorController tor_status;

  public ButtonTorController() {
    this.set_label("Tor status");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
  }

  public void on_click_button() {
    if (this.tor_status == null || !this.tor_status.get_visible()) {
      this.tor_status = new DialogTorController();
      this.tor_status.show_dialog();
    }
  }

  public void set_surf_activated() {
    this.set_sensitive(true);
  }

  public void set_surf_deactivated() {
    this.set_sensitive(false);
  }
}
