using Gtk;


public class LabelBootStatus: Label {
  public LabelBootStatus() {
    this.set_label("Boot status");
  }
}


public class ButtonEnableBoot: Switch {
  public ThreadPool<ThreadEnableAnonSurf> pool;
  private bool updating_state;

  public ButtonEnableBoot() {
    this.state_set.connect((state) => {
      if (!this.updating_state) {
        if (state) {
          on_enable_boot();
        } else {
          on_disable_boot();
        }
      }
      return false;
    });

    try {
      this.pool = new ThreadPool<ThreadEnableAnonSurf>.with_owned_data((worker) => {
        worker.run();
      }, 1, false);
    } catch (GLib.Error error) {
      send_notification("Thread error", error.message, NotifyLevel.Error);
    }
  }

  private void on_enable_boot() {
    try {
      this.pool.add(new ThreadEnableAnonSurf(true));
    } catch (ThreadError error) {
      send_notification("Thread error", error.message, NotifyLevel.Error);
    }
  }

  private void on_disable_boot() {
    try {
      this.pool.add(new ThreadEnableAnonSurf(false));
    } catch (ThreadError error) {
      send_notification("Thread error", error.message, NotifyLevel.Error);
    }
  }

  public void update_label(bool is_enabled_boot) {
    if (this.pool.get_num_threads() > 0) {
      return;
    }

    this.updating_state = true;
    this.set_state(is_enabled_boot);
    this.updating_state = false;
  }
}


public class BoxBootStatus: Box {
  public LabelBootStatus label_boot_status;
  public ButtonEnableBoot button_enable_boot;

  public BoxBootStatus() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    create_box_layout();
  }

  private void create_box_layout() {
    label_boot_status = new LabelBootStatus();
    button_enable_boot = new ButtonEnableBoot();

    this.pack_start(label_boot_status, true, false, 3);
    this.pack_start(button_enable_boot, true, false, 3);
  }

  public void update_boot_status() {
    button_enable_boot.update_label(systemd_unit_matches("is-enabled", "anonsurfd.service"));
  }
}
