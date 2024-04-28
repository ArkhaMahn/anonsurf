using Gtk;

// TODO: if start supprocess, it's better to use absolute path of cmd service (check path exists first)
public class BoxMainLayout: Box {
  private BoxBootStatus box_boot_status;
  public BoxButtonArea boxButton;

  public BoxMainLayout() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    box_boot_status = new BoxBootStatus();
    boxButton = new BoxButtonArea();

    this.pack_start(box_boot_status, true, true, 3);
    this.pack_start(boxButton, false, true, 3);
  }

  // private bool check_tor_running() {
  //   File file = File.new_for_path("/run/systemd/units/invocation:tor.service");
  //   return file.query_exists();
  // }

  public void update_layout(bool is_surf_running) {
    box_boot_status.update_boot_status();
    boxButton.refresh_box(is_surf_running);
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
    bool is_surf_running = check_surf_running();
    this.sys_tray.on_update_systray(is_surf_running);

    if (this.window.title_bar.get_visible()) {
      this.window.main_layout.update_layout(is_surf_running);
      this.window.title_bar.on_update_title_bar(is_surf_running);
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
