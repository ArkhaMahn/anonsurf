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


public class BoxMainLayout: Box {
  private BoxStatusBoot boxBoot;
  private BoxButtonArea boxButton;

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
