/*
Create area that shows current status if AnonSurf is enabled with system (via label)
And a button to enable, disable it at boot
*/
using Gtk;


public class LabelBootStatus: Label {
  public LabelBootStatus() {
    this.set_label("Boot status");
  }
}


public class ButtonEnableBoot: Switch {
  public ThreadPool<ThreadEnableAnonSurf> pool;

  public ButtonEnableBoot() {
    this.notify["active"].connect(() => {
      /*
        When "activate" signal is sent, there are multiple actual signals here
        By debug, normal fresher gives signal 128 or the other. Meanwhile, click
        on Switch object gives signal either 2178 or 130.
        - Signal 2178: Switch from Disable to enable
        - Signal 130: Switch from Enable to disable
        From this.get_state_flags().to_string(), both values are GTK_STATE_FLAG_PRELIGHT
      */
      switch (this.get_state_flags()) {
        case 2178: // Click enable
          on_enable_boot();
          break;
        case 130: // Click disable
          on_disable_boot();
          break;
        default:
          break;
      }
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
      // Do not update status when a thread is running
      // Otherwise subprocess will be called multiple times
      return;
    }

    if (is_enabled_boot == true) {
      this.set_state(true);
    } else {
      this.set_state(false);
    }
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
    this.pack_start(button_enable_boot, false, false, 3);
  }

  private bool check_multi_user_unit() {
    File file = File.new_for_path("/etc/systemd/system/multi-user.target.wants/anonsurfd.service");
    return file.query_exists();
  }

  private bool check_default_target_unit() {
    File file = File.new_for_path("/etc/systemd/system/default.target.wants/anonsurfd.service");
    return file.query_exists();
  }

  public void update_boot_status() {
    if (check_multi_user_unit() == false && check_default_target_unit() == false){
      button_enable_boot.update_label(false);
    } else {
      button_enable_boot.update_label(true);
    }
  }
}
