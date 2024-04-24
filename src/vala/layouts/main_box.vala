using Gtk;
using Soup;

const uchar BUTTON_SIZE_X = 105;
const uchar BUTTON_SIZE_Y = 65;
const uchar BUTTON_BOOT_SIZE_X = 60;
const uchar BUTTON_BOOT_SIZE_Y = 48;

public enum AnonSurfStatusErr {
  OK,
  TORNOTRUNNING,
  ANONSURFNOTRUNNING
}

/*
  TODO: change classes to private if possible
*/


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
    GLib.Object(orientation: Orientation.HORIZONTAL);
    create_box_layout();
    refresh_box();
  }

  private void create_box_layout() {
    label_boot = new LabelBootStatus();
    btn_boot = new ButtonBoot();

    this.pack_start(label_boot, false, false, 3);
    this.pack_start(btn_boot, false, false, 3);
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


public class ButtonStart: Button {
  public ThreadPool<WorkerSystemdService> pool;

  public ButtonStart() {
    this.set_label("Start");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
    this.pool = new ThreadPool<WorkerSystemdService>.with_owned_data ((worker) => {
      worker.run();
    }, 1, false);
    // 1 -> Max threads in pool
  }

  private void on_click_button() {
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
    this.clicked.connect(on_click_myip);
    this.pool = new ThreadPool<WorkerCheckTor>.with_owned_data ((worker) => {
      worker.run ();
    }, 1, false);
  }

  private void on_click_myip() {
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


public class BoxTopButtons: Box {
  public ButtonStart btnStart;
  public ButtonMyIP btnMyIP;

  public BoxTopButtons() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    btnStart = new ButtonStart();
    btnMyIP = new ButtonMyIP();
    this.pack_start(btnStart, false, true, 3);
    this.pack_start(btnMyIP, false, true, 3);
  }

  public void refresh_box(bool is_surf_running) {
    btnStart.update_button(is_surf_running);
    btnMyIP.update_button();
  }
}


public class ButtonChangeID: Button {
  public ButtonChangeID() {
    this.set_label("Change ID");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(change_id);
  }

  private void change_id() {
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
    this.clicked.connect(check_tor_connection);
  }

  private void check_tor_connection() {
    this.tor_status.show_dialog();
    // this.tor_status.run();
    // this.tor_status.destroy();
  }

  public void update_button(bool is_surf_running) {
    if (is_surf_running) {
      this.set_sensitive(true);
    } else {
      this.set_sensitive(false);
    }
  }
}


public class BoxBottomButtons: Box {
  public ButtonChangeID btnChangeID;
  public ButtonStatus btnStatus;

  public BoxBottomButtons() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    btnChangeID = new ButtonChangeID();
    btnStatus = new ButtonStatus();
    this.pack_start(btnStatus, false, true, 3);
    this.pack_start(btnChangeID, false, true, 3);
  }

  public void refresh_box(bool is_surf_running) {
    btnChangeID.update_button(is_surf_running);
    btnStatus.update_button(is_surf_running);
  }
}


public class BoxButtonArea: Box {
  public BoxBottomButtons boxBottom;
  public BoxTopButtons boxTop;

  public BoxButtonArea() {
    GLib.Object(orientation: Orientation.VERTICAL);
    boxBottom = new BoxBottomButtons();
    boxTop = new BoxTopButtons();
    this.pack_start(boxTop, false, true, 3);
    this.pack_start(boxBottom, false, true, 3);
  }

  public void refresh_box(bool is_surf_running) {
    boxBottom.refresh_box(is_surf_running);
    boxTop.refresh_box(is_surf_running);
  }
}


public class BoxMainLayout: Box {
  public BoxStatusBoot boxBoot;
  public BoxButtonArea boxButton;

  public BoxMainLayout() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    boxBoot = new BoxStatusBoot();
    boxButton = new BoxButtonArea();
    this.pack_start(boxBoot, false, true, 3);
    this.pack_start(boxButton, false, true, 3);
    update_layout();
  }

  // private bool check_tor_running() {
  //   File file = File.new_for_path("/run/systemd/units/invocation:tor.service");
  //   return file.query_exists();
  // }

  private bool check_surf_running() {
    File file = File.new_for_path("/run/systemd/units/invocation:anonsurfd.service");
    return file.query_exists();
  }

  public void update_layout() {
    bool is_surf_running = check_surf_running();

    boxBoot.refresh_box();
    boxButton.refresh_box(is_surf_running);
  }
}
