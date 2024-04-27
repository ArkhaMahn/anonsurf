using Gtk;

const uchar BUTTON_SIZE_X = 105;
const uchar BUTTON_SIZE_Y = 65;
const uchar BUTTON_BOOT_SIZE_X = 60;
const uchar BUTTON_BOOT_SIZE_Y = 48;

public enum AnonSurfStatusErr {
  OK,
  TORNOTRUNNING,
  ANONSURFNOTRUNNING
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
  private BoxStatusBoot boxBoot;
  public BoxButtonArea boxButton;

  public BoxMainLayout() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    boxBoot = new BoxStatusBoot();
    boxButton = new BoxButtonArea();
    this.pack_start(boxBoot, true, true, 3);
    this.pack_start(boxButton, false, true, 3);
  }

  // private bool check_tor_running() {
  //   File file = File.new_for_path("/run/systemd/units/invocation:tor.service");
  //   return file.query_exists();
  // }

  public void update_layout(bool is_surf_running) {
    boxBoot.refresh_box();
    boxButton.refresh_box(is_surf_running);
  }
}
