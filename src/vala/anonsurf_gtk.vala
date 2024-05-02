using Gtk;

// TODO: if start supprocess, it's better to use absolute path of cmd service (check path exists first)
public class BoxMainLayout: Box {
  private BoxBootStatus box_boot_status;
  public BoxButtonControl box_button_control;

  public BoxMainLayout() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    this.box_boot_status = new BoxBootStatus();
    this.box_button_control = new BoxButtonControl();

    this.pack_start(this.box_boot_status, true, true, 3);
    this.pack_start(this.box_button_control, false, true, 3);
  }

  // private bool check_tor_running() {
  //   File file = File.new_for_path("/run/systemd/units/invocation:tor.service");
  //   return file.query_exists();
  // }
  public void update_boot_status() {
    // TODO complete here
  }

  public void set_surf_activated() {
    this.box_button_control.set_surf_activated();
  }

  public void set_surf_deactivated() {
    this.box_button_control.set_surf_deactivated();
  }
}


public class AnonSurfWindow: Gtk.Window {
  public BoxMainLayout main_layout;
  public AnonSurfTitleBar title_bar;

  public AnonSurfWindow() {
    this.title_bar = new AnonSurfTitleBar();
    this.main_layout = new BoxMainLayout();

    this.set_titlebar(this.title_bar);
    this.set_icon_name("anonsurf");
    this.set_resizable(false);
    this.add(this.main_layout);
  }
}


public class AnonSurfApp: GLib.Application {
  public AnonSurfWindow window;
  private AnonSurfSystrayIcon sys_tray;

  public AnonSurfApp() {
    Object(application_id: "org.parrot.anonsurf.gui");
  }

  public override void activate() {
    this.window = new AnonSurfWindow();
    this.refresh_gui();
    this.window.show_all();

    Timeout.add(200, refresh_gui);
    Gtk.main();
  }

  public void set_systray_connection(AnonSurfSystrayIcon systray) {
    this.sys_tray = systray;
  }

  private bool check_surf_running() {
    File file = File.new_for_path("/run/systemd/units/invocation:anonsurfd.service");
    return file.query_exists();
  }

  private bool refresh_gui() {
    /*
    Update system tray
    If visible:
    - Update title bar and main layout
    */

    /*
      TODO
      if is_surf_running:
        update_systray
        if not get_visible:
          return true
        update_boot_status
        update_titlebar
        update_layout
      else:
        update_systray
        if not get_visible:
          return true
        update_boot_status
        update_titlebar
        update_layout
      return true
    */
    // bool is_surf_running = check_surf_running();
    // this.sys_tray.update_menu_and_icon(is_surf_running);
    // // TODO optimize conditions to avoid same code in multiple small branches

    // if (this.window.title_bar.get_visible()) {
    //   this.window.main_layout.update_layout(is_surf_running);
    //   this.window.title_bar.on_update_title_bar(is_surf_running);
    // }
    bool is_surf_running = check_surf_running();
    if (is_surf_running) {
      this.sys_tray.set_surf_activated();

      if (!this.window.title_bar.get_visible()) {
        return true;
      }

      // TODO thing about checking tor's conditions too
      this.window.main_layout.set_surf_activated();
      this.window.title_bar.set_surf_activated();

    } else {
      this.sys_tray.set_surf_deactivated();
      if (!this.window.title_bar.get_visible()) {
        return true;
      }
      // TODO thing about checking tor's conditions too
      this.window.main_layout.set_surf_deactivated();
      this.window.title_bar.set_surf_deactivated();
    }

    return true;
  }
}


int main(string[] args) {
  Gtk.init(ref args);

  AnonSurfApp app = new AnonSurfApp();
  AnonSurfSystrayIcon sys_tray = new AnonSurfSystrayIcon();

  sys_tray.set_app_connection(app);
  app.set_systray_connection(sys_tray);

  Notify.init("AnonSurf GTK");
  return app.run(args);
}
