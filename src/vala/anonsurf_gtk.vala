using Gtk;


public class AnonSurfLayout: Gtk.Window {
  //   TODO: update icon of tray icon by status
  // TODO: update title bar by status. Maybe move status area to this
  public BoxMainLayout main_layout;

  public AnonSurfLayout(AnonSurfDialogOptions dialog_options) {
    this.main_layout = new BoxMainLayout();
    this.set_titlebar(new AnonSurfTitleBar(dialog_options));
    this.set_icon_name("anonsurf");
    this.set_resizable(false);
    this.add(this.main_layout);
  }
}


public class AnonSurfApp: GLib.Application {
  private AnonSurfDialogOptions dialog_options;
  private AnonSurfLayout layout;

  public AnonSurfApp(AnonSurfDialogOptions dialog_options) {
    Object(application_id: "org.parrot.anonsurf-gtk");
    this.dialog_options = dialog_options;
  }

  public override void activate() {
    this.layout = new AnonSurfLayout(this.dialog_options);
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

  var dialog_options = new AnonSurfDialogOptions();
  var app = new AnonSurfApp(dialog_options);
  var sys_tray = new AnonSurfStatusIcon(app, dialog_options);
  Notify.init("AnonSurf GTK");
  return app.run(args);
}
