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
  TODO: update icon of tray icon by status
  TODO: update title bar by status. Maybe move status area to this
  TODO: add status
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
  public ButtonBoot() {
    this.set_size_request(BUTTON_BOOT_SIZE_X, BUTTON_BOOT_SIZE_Y);
    this.clicked.connect(on_click_boot_action);
  }

  private void on_click_boot_action() {
    if (this.get_label() == "Enable") {
      print("Enabling boot\n");
    }
    else {
      print("Disabling boot\n");
    }
  }

  public void on_update_label(bool is_enabled_boot) {
    if (is_enabled_boot == true) {
      this.set_label("Disable");
    } else {
      this.set_label("Enable");
    }
  }
}


public class BoxStatusBoot: Box {
  public LabelBootStatus label_boot;
  public ButtonBoot btn_boot;

  public BoxStatusBoot() {
    GLib.Object(orientation: Orientation.VERTICAL);
    create_box_layout();
    refresh();
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

  public void refresh() {
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


public class LabelAnonSurfStatus: Label {
  public LabelAnonSurfStatus(AnonSurfStatusErr anonsurf_status) {
    on_update_label(anonsurf_status);
  }

  public void on_update_label(AnonSurfStatusErr anonsurf_status) {
    if (anonsurf_status == AnonSurfStatusErr.TORNOTRUNNING) {
      this.set_label("Tor died");
    } else if (anonsurf_status == AnonSurfStatusErr.ANONSURFNOTRUNNING) {
      this.set_label("Deactivated");
    } else {
      this.set_label("Activated");
    }
  }
}


public class ButtonAnonSurfAction: Button {
  public ButtonAnonSurfAction(bool anonsurf_status) {
    on_update_button(anonsurf_status);
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_anonsurf_action);
  }

  private void on_click_anonsurf_action() {
    if (this.get_label() == "Start") {
      print("Starting\n");
    }
    else {
      print("Stopping\n");
    }
  }

  public void on_update_button(bool anonsurf_status) {
    if (anonsurf_status == true) {
      this.set_label("Stop");
    } else {
      this.set_label("Start");
    }
  }
}


public class ButtonChangeID: Button {
  public ButtonChangeID(bool anonsurf_status) {
    this.set_label("Change ID");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_changeid);
    on_update_button(anonsurf_status);
  }

  // TODO require control port from setting
  // TODO socket with vala
  // TODO create a thread

  private void on_click_changeid() {
    print("Changing id\n");
  }

  public void on_update_button(bool anonsurf_status) {
    if (anonsurf_status == true) {
      this.set_sensitive(true);
    } else {
      this.set_sensitive(false);
    }
  }
}


public class ButtonRestart: Button {
  public ButtonRestart(bool anonsurf_status) {
    this.set_label("Restart");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_restart);

    on_update_button(anonsurf_status);
  }

  private void on_click_restart() {
    print("Restarting\n");
  }

  public void on_update_button(bool anonsurf_status) {
    if (anonsurf_status == true) {
      this.set_sensitive(true);
    } else {
      this.set_sensitive(false);
    }
  }
}


public class ButtonMyIP: Button {
  public ButtonMyIP() {
    this.set_label("My IP");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_myip);
  }

  private bool http_tor_status() {
    /*
    https://wiki.gnome.org/Projects/Vala/LibSoupSample
    Sorry. You are not using Tor.
    */
    var session = new Soup.Session();
    var message = new Soup.Message("GET", "https://check.torproject.org/");
    session.send_message(message);
    // TODO Reverse this condition: Maybe error happens. Check for success, error and not using tor?
    if (((string)message.response_body.data).contains("Sorry. You are not using Tor.")) {
      return false;
    }
    return true;
  }

  private void check_tor_status() {
    // TODO show notification instead
    if (http_tor_status()) {
      print("Using Tor\n");
    } else {
      print("Not using Tor\n");
    }
  }

  private void on_click_myip() {
    // FIXME button wasn't disabled
    this.set_sensitive(false);
    var background_check = new Thread<void>("background_check", check_tor_status);
    this.set_sensitive(true);
  }
}


public class ImageAnonSurfStatus: Image {
  public ImageAnonSurfStatus(AnonSurfStatusErr anonsurf_status) { // TODO get status
    on_update_image(anonsurf_status);
  }

  public void on_update_image(AnonSurfStatusErr anonsurf_status) {
    if (anonsurf_status == AnonSurfStatusErr.TORNOTRUNNING) {
      this.set_from_icon_name("security-low", DIALOG);
    } else if (anonsurf_status == AnonSurfStatusErr.ANONSURFNOTRUNNING) {
      this.set_from_icon_name("security-medium", DIALOG);
    } else {
      this.set_from_icon_name("security-high", DIALOG);
    }
  }
}


public class MainLayout: Box {
  // private LabelBootStatus label_boot_status;
  //  private LabelAnonSurfStatus label_anonsurf_status;
  private ButtonAnonSurfAction button_anonsurf_actions; // Start or Stop anonsurf
  private ButtonChangeID button_change_id;
  private ButtonMyIP button_my_ip;
  private ButtonRestart button_restart;
  // private ButtonBootAction button_boot_actions;
  //  private ImageAnonSurfStatus image_status;

  private BoxStatusBoot box_boot;

  private Box box_detail_area;
  private Box box_button_area;

  private Frame frame_boot_area;


  public MainLayout() {
    GLib.Object(orientation: Orientation.HORIZONTAL);

    init_all_objects();

    var box_boot_info = new Box(Orientation.VERTICAL, 3);

    //  box_extra_details.pack_start(box_detail_area, true, true, 3);
    box_boot_info.pack_start(frame_boot_area, false, true, 3);

    this.pack_start(box_boot_info, false, true, 3);
    this.pack_start(box_button_area, false, true, 1);
  }

  private void init_all_objects() {
    create_object_with_status();
    //  init_detail_area();
    init_boot_area();
    init_button_area();
  }

  //  private void init_detail_area() {
  //    box_detail_area = new Box(Orientation.HORIZONTAL, 3);

  //    box_detail_area.pack_start(this.image_status, false, true, 3);
  //    box_detail_area.pack_start(this.label_anonsurf_status, false, true, 3);
  //  }

  private void init_boot_area() {
    box_boot = new BoxStatusBoot();
    frame_boot_area = new Frame("Boot status");

    // box_boot_area.pack_start(this.label_boot_status, false, true, 3);
    // box_boot_area.pack_start(this.button_boot_actions, false, true, 3);

    frame_boot_area.add(box_boot);
  }

  private void init_button_area() {
    box_button_area = new Box(Orientation.VERTICAL, 3);

    var box_button_upper = new Box(Orientation.HORIZONTAL, 3);
    var box_button_lower = new Box(Orientation.HORIZONTAL, 3);

    box_button_upper.pack_start(this.button_anonsurf_actions, false, true, 3);
    box_button_upper.pack_start(this.button_restart, false, true, 3);
    box_button_lower.pack_start(this.button_change_id, false, true, 3);
    box_button_lower.pack_start(this.button_my_ip, false, true, 3);

    box_button_area.pack_start(box_button_upper, false, true, 3);
    box_button_area.pack_start(box_button_lower, false, true, 3);
  }

  private void create_object_with_status() {
    bool anonsurf_status = is_anonsurf_running();
    bool tor_status = is_tor_running();
    AnonSurfStatusErr surf_run_status;

    // this.label_boot_status = new LabelBootStatus();
    this.button_anonsurf_actions = new ButtonAnonSurfAction(anonsurf_status);
    this.button_change_id = new ButtonChangeID(anonsurf_status);
    this.button_restart = new ButtonRestart(anonsurf_status);
    // this.button_boot_actions = new ButtonBootAction();
    this.button_my_ip = new ButtonMyIP();

    if (anonsurf_status == true) {
      if (tor_status == true) {
        surf_run_status = AnonSurfStatusErr.OK;
      } else {
        surf_run_status = AnonSurfStatusErr.TORNOTRUNNING;
      }
    } else {
      surf_run_status = AnonSurfStatusErr.ANONSURFNOTRUNNING;
    }

    //  this.image_status = new ImageAnonSurfStatus(surf_run_status);
    //  this.label_anonsurf_status = new LabelAnonSurfStatus(surf_run_status);
  }

  public void on_update_layout() {
    // TODO call refresh method of BoxStatusBoot
    bool anonsurf_status = is_anonsurf_running();
    bool tor_status = is_tor_running();
    AnonSurfStatusErr surf_run_status;

    // this.label_boot_status.refresh();
    // this.button_boot_actions.on_update_button(anonsurf_boot_status);

    this.button_anonsurf_actions.on_update_button(anonsurf_status);
    this.button_change_id.on_update_button(anonsurf_status);
    this.button_restart.on_update_button(anonsurf_status);

    if (anonsurf_status == true) {
      if (tor_status == true) {
        surf_run_status = AnonSurfStatusErr.OK;
      } else {
        surf_run_status = AnonSurfStatusErr.TORNOTRUNNING;
      }
    } else {
      surf_run_status = AnonSurfStatusErr.ANONSURFNOTRUNNING;
    }

    //  this.label_anonsurf_status.on_update_label(surf_run_status);
    //  this.image_status.on_update_image(surf_run_status);
  }
}
