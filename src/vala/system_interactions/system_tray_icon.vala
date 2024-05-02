/*
Create a system tray icon in the notification area
This sytem tray icon will update icon with AnonSurf's status
*/

using Gtk;


public class AnonSurfSystrayIcon: StatusIcon {
  private AnonSurfSystrayMenu menu;
  private AnonSurfApp app;

  public AnonSurfSystrayIcon() {
    this.menu = new AnonSurfSystrayMenu();

    this.activate.connect(on_left_click_systray);
    this.popup_menu.connect(on_right_click_systray);
  }

  private void on_right_click_systray(uint button, uint time) {
    this.menu.popup_at_pointer();
  }

  private void on_left_click_systray() {
    this.app.activate();
  }

  public void set_app_connection(AnonSurfApp app) {
    this.app = app;
    this.menu.set_app_connection(app);
  }

  public void set_surf_activated() {
    this.set_from_icon_name("security-high");
    this.menu.set_surf_activated();
  }

  public void set_surf_deactivated() {
    this.set_from_icon_name("anonsurf");
    this.menu.set_surf_deactivated();
  }
}
