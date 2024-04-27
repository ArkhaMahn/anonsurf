using Gtk;


public class ButtonStart: Button {
  public ThreadPool<WorkerSystemdService> pool;

  public ButtonStart() {
    this.set_label("Start");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
    try {
      this.pool = new ThreadPool<WorkerSystemdService>.with_owned_data ((worker) => {
        worker.run();
      }, 1, false);
    } catch (GLib.Error error) {
      print("Failed to create thread: %s\n", error.message);
    }
  }

  public void on_click_button() {
    if (this.pool.get_num_threads() > 0) {
      return;
    }

    try {
      if (this.get_label() == "Start") {
        this.pool.add(new WorkerSystemdService(true));
      } else {
        this.pool.add(new WorkerSystemdService(false));
      }
    } catch (ThreadError e) {
      print("ThreadError: %s\n", e.message);
    }
  }

  public void update_button(bool is_surf_running) {
    if (this.pool.get_num_threads() > 0) {
      this.set_sensitive(false);
    } else {
      this.set_sensitive(true);
      if (is_surf_running) {
        this.set_label("Stop");
      } else {
        this.set_label("Start");
      }
    }
  }
}


public class ButtonMyIP: Button {
  private ThreadPool<WorkerCheckTor> pool;

  public ButtonMyIP() {
    this.set_label("My IP");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
    try {
      this.pool = new ThreadPool<WorkerCheckTor>.with_owned_data ((worker) => {
        worker.run ();
      }, 1, false);
    } catch (GLib.Error error) {
      print("Failed to create thread: %s\n", error.message);
    }
  }

  public void on_click_button() {
    if (this.pool.get_num_threads() > 0) {
      return;
    }
    try {
      this.pool.add(new WorkerCheckTor());
    } catch (ThreadError e) {
      print("ThreadError: %s\n", e.message);
    }
  }

  public void update_button() {
    if (this.pool.get_num_threads() > 0) {
      this.set_sensitive(false);
    } else {
      this.set_sensitive(true);
    }
  }
}


public class ButtonChangeID: Button {
  public ButtonChangeID() {
    this.set_label("Change ID");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
  }

  public void on_click_button() {
    // TODO changeid with socket. Value of port parses from config file
  }

  public void update_button(bool is_surf_running) {
    if (is_surf_running) {
      this.set_sensitive(true);
    } else {
      this.set_sensitive(false);
    }
  }
}


public class ButtonStatus: Button {
  private StatusDialog tor_status;

  public ButtonStatus() {
    this.tor_status = new StatusDialog();
    this.set_label("Tor stat");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
  }

  public void on_click_button() {
    this.tor_status.show_dialog();
  }

  public void update_button(bool is_surf_running) {
    if (is_surf_running) {
      this.set_sensitive(true);
    } else {
      this.set_sensitive(false);
    }
  }
}
