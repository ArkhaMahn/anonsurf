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
  public ButtonStart button_start;
  public ButtonMyIP button_my_ip;

  public BoxTopButtons() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    this.button_start = new ButtonStart();
    this.button_my_ip = new ButtonMyIP();

    this.pack_start(button_start, false, true, 12);
    this.pack_start(button_my_ip, false, true, 10);
  }

  public void set_surf_activated() {
    this.button_start.set_surf_activated();
    this.button_start.check_thread_is_running();
    this.button_my_ip.check_thread_is_running();
  }

  public void set_surf_deactivated() {
    this.button_start.set_surf_deactivated();
    this.button_start.check_thread_is_running();
    this.button_my_ip.check_thread_is_running();
  }
}

public class BoxBottomButtons: Box {
  public ButtonChangeID button_change_id;
  public ButtonTorController button_tor_controller;

  public BoxBottomButtons() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    this.button_change_id = new ButtonChangeID();
    this.button_tor_controller = new ButtonTorController();

    this.pack_start(this.button_tor_controller, false, true, 12);
    this.pack_start(this.button_change_id, false, true, 10);
  }

  public void set_surf_activated() {
    this.button_change_id.set_surf_activated();
    this.button_tor_controller.set_surf_activated();
  }

  public void set_surf_deactivated() {
    this.button_change_id.set_surf_deactivated();
    this.button_tor_controller.set_surf_deactivated();
  }
}

public class BoxButtonControl: Box {
  public BoxBottomButtons box_bottom_buttons;
  public BoxTopButtons box_top_buttons;

  public BoxButtonControl() {
    GLib.Object(orientation: Orientation.VERTICAL);
    this.box_bottom_buttons = new BoxBottomButtons();
    this.box_top_buttons = new BoxTopButtons();

    this.pack_start(this.box_top_buttons, false, true, 3);
    this.pack_start(this.box_bottom_buttons, false, true, 3);
  }

  public void set_surf_activated() {
    this.box_top_buttons.set_surf_activated();
    this.box_bottom_buttons.set_surf_activated();
  }

  public void set_surf_deactivated() {
    this.box_top_buttons.set_surf_deactivated();
    this.box_bottom_buttons.set_surf_deactivated();
  }
}
