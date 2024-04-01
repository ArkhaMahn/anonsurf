using Gtk;

// TODO: if start supprocess, it's better to use absolute path of cmd service (check path exists first)


public class AnonSurfLayout: Gtk.Window {
  //   TODO: update icon of tray icon by status
  // TODO: update title bar by status. Maybe move status area to this
  public BoxMainLayout main_layout;

  public AnonSurfLayout() {
    this.main_layout = new BoxMainLayout();
    this.set_titlebar(new AnonSurfTitleBar());
    this.set_icon_name("anonsurf");
    this.set_resizable(false);
    this.add(this.main_layout);
  }
}


public class AnonSurfApp: GLib.Application {
  private AnonSurfLayout layout;

  public AnonSurfApp() {
    Object(application_id: "org.parrot.anonsurf.gui");
  }

  public override void activate() {
    this.layout = new AnonSurfLayout();
    this.layout.show_all();
    Timeout.add(200, refresh_gui);
    Gtk.main();
  }

  private bool refresh_gui() {
    this.layout.main_layout.update_layout();
    return true;
  }
}


int main(string[] args) {
  Gtk.init(ref args);

  var app = new AnonSurfApp();
  var sys_tray = new AnonSurfStatusIcon(app);
  Notify.init("AnonSurf GTK");
  return app.run(args);
}
