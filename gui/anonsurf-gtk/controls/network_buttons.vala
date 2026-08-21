using Gtk;

public class ButtonStart: Button {
  public ThreadPool<ThreadStartAnonSurf> pool;
  private DialogAskKill dialog_ask_kill;

  public ButtonStart() {
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
    try {
      this.pool = new ThreadPool<ThreadStartAnonSurf>.with_owned_data((worker) => {
        worker.run();
      }, 1, false);
    } catch (GLib.Error error) {
      send_notification("Thread error", error.message, NotifyLevel.Error);
    }
  }

  public void on_click_button() {
    if (this.pool.get_num_threads() > 0) {
      return;
    }

    this.dialog_ask_kill = new DialogAskKill();
    dialog_ask_kill.start_dialog();

    this.dialog_ask_kill.destroy.connect(() => {
      if (!this.dialog_ask_kill.should_continue) {
        return;
      }
      try {
        if (this.get_label() == "Start") {
          this.pool.add(new ThreadStartAnonSurf(true));
        } else {
          this.pool.add(new ThreadStartAnonSurf(false));
        }
      } catch (ThreadError error) {
        send_notification("Error while stopping AnonSurf", error.message, NotifyLevel.Error);
      }
    });
  }

  public void check_thread_is_running() {
    if (this.pool.get_num_threads() > 0) {
      this.set_sensitive(false);
    } else {
      this.set_sensitive(true);
    }
  }

  public void set_surf_activated() {
    this.set_label("Stop");
  }

  public void set_surf_deactivated() {
    this.set_label("Start");
  }
}

public class ButtonMyIP: Button {
  private ThreadPool<ThreadMyIP> pool;

  public ButtonMyIP() {
    this.set_label("My IP");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
    try {
      this.pool = new ThreadPool<ThreadMyIP>.with_owned_data((worker) => {
        worker.run();
      }, 1, false);
    } catch (GLib.Error error) {
      send_notification("Thread error", error.message, NotifyLevel.Error);
    }
  }

  public void on_click_button() {
    if (this.pool.get_num_threads() > 0) {
      return;
    }
    try {
      this.pool.add(new ThreadMyIP());
    } catch (ThreadError error) {
      send_notification("Thread error", error.message, NotifyLevel.Error);
    }
  }

  public void check_thread_is_running() {
    if (this.pool.get_num_threads() > 0) {
      this.set_sensitive(false);
    } else {
      this.set_sensitive(true);
    }
  }
}
