using Gtk;

// TODO: if start supprocess, it's better to use absolute path of cmd service (check path exists first)


public class AnonSurfWindow: Gtk.Window {
  //   TODO: update icon of tray icon by status
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
  private AnonSurfWindow layout;
  private AnonSurfSystrayIcon sys_tray;

  public AnonSurfApp() {
    Object(application_id: "org.parrot.anonsurf.gui");
  }

  public override void activate() {
    this.layout = new AnonSurfWindow();
    this.refresh_gui();
    this.layout.show_all();

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

    this.layout.main_layout.update_layout(is_surf_running);
    this.sys_tray.on_update_systray(is_surf_running);
    this.layout.title_bar.on_update_title_bar(is_surf_running);
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
