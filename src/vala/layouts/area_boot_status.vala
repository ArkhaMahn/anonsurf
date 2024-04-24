using Gtk;

public class LabelBootStatus: Label {
  public LabelBootStatus() {
  }

  public void on_update_label(bool is_enabled_boot) {
    if (is_enabled_boot == false) {
      this.set_label("Disabled");
    } else {
      this.set_label("Enabled");
    }
  }
}


public class ButtonBoot: Button {
  public ThreadPool<WorkerSystemdStartup> pool;

  public ButtonBoot() {
    this.set_size_request(BUTTON_BOOT_SIZE_X, BUTTON_BOOT_SIZE_Y);
    this.clicked.connect(on_click_boot_action);
    this.pool = new ThreadPool<WorkerSystemdStartup>.with_owned_data ((worker) => {
      worker.run();
    }, 1, false);
  }

  private void on_click_boot_action() {
    try {
      if (this.get_label() == "Enable") {
        this.pool.add(new WorkerSystemdStartup(true));
      } else {
        this.pool.add(new WorkerSystemdStartup(false));
      }
    } catch (ThreadError e) {
      print("ThreadError: %s\n", e.message);
    }
  }

  public void on_update_label(bool is_enabled_boot) {
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
  public LabelBootStatus label_boot;
  public ButtonBoot btn_boot;

  public BoxStatusBoot() {
    GLib.Object(orientation: Orientation.VERTICAL);
    create_box_layout();
    refresh_box();
  }

  private void create_box_layout() {
    label_boot = new LabelBootStatus();
    btn_boot = new ButtonBoot();

    this.pack_start(label_boot, true, true, 3);
    this.pack_start(btn_boot, true, false, 3);
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
      label_boot.on_update_label(false);
      btn_boot.on_update_label(false);
      // TODO update button
      // TODO update label
    } else {
      label_boot.on_update_label(true);
      btn_boot.on_update_label(true);
    }
  }
}
