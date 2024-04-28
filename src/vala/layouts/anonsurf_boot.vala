/*
Create area that shows current status if AnonSurf is enabled with system (via label)
And a button to enable, disable it at boot
*/
using Gtk;


public class LabelBootStatus: Label {
  public LabelBootStatus() {
  }

  public void update_label(bool is_enabled_boot) {
    if (is_enabled_boot == false) {
      this.set_label("Disabled");
    } else {
      this.set_label("Enabled");
    }
  }
}


public class ButtonEnableBoot: Button {
  public ThreadPool<ThreadEnableAnonSurf> pool;

  public ButtonEnableBoot() {
    this.set_size_request(BUTTON_BOOT_SIZE_X, BUTTON_BOOT_SIZE_Y);
    this.clicked.connect(on_click_boot_action);
    try {
      this.pool = new ThreadPool<ThreadEnableAnonSurf>.with_owned_data ((worker) => {
        worker.run();
      }, 1, false);
    } catch (GLib.Error error) {
      print("Failed to create thread: %s\n", error.message);
    }
  }

  private void on_click_boot_action() {
    try {
      if (this.get_label() == "Enable") {
        this.pool.add(new ThreadEnableAnonSurf(true));
      } else {
        this.pool.add(new ThreadEnableAnonSurf(false));
      }
    } catch (ThreadError e) {
      print("ThreadError: %s\n", e.message);
    }
  }

  public void update_label(bool is_enabled_boot) {
    if (this.pool.get_num_threads() > 0) {
      this.set_sensitive(false);
    } else {
      this.set_sensitive(true);
      if (is_enabled_boot == true) {
        this.set_label("Disable");
      } else {
        this.set_label("Enable");
      }
    }
  }
}


public class BoxStatusBoot: Box {
  public LabelBootStatus label_boot_status;
  public ButtonEnableBoot button_enable_boot;

  public BoxStatusBoot() {
    GLib.Object(orientation: Orientation.VERTICAL);
    create_box_layout();
    refresh_box();
  }

  private void create_box_layout() {
    label_boot_status = new LabelBootStatus();
    button_enable_boot = new ButtonEnableBoot();

    this.pack_start(label_boot_status, true, true, 3);
    this.pack_start(button_enable_boot, true, true, 3);
  }

  private bool check_multi_user_unit() {
    File file = File.new_for_path("/etc/systemd/system/multi-user.target.wants/anonsurfd.service");
    return file.query_exists();
  }

  private bool check_default_target_unit() {
    File file = File.new_for_path("/etc/systemd/system/default.target.wants/anonsurfd.service");
    return file.query_exists();
  }

  public void refresh_box() {
    if (check_multi_user_unit() == false && check_default_target_unit() == false){
      label_boot_status.update_label(false);
      button_enable_boot.update_label(false);
    } else {
      label_boot_status.update_label(true);
      button_enable_boot.update_label(true);
    }
  }
}
