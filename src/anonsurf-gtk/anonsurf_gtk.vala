using Gtk;


public class BoxMainLayout: Box {
  public BoxBootStatus box_boot_status;
  public BoxButtonControl box_button_control;

  public BoxMainLayout() {
    GLib.Object(orientation: Orientation.VERTICAL);
    this.box_boot_status = new BoxBootStatus();
    this.box_button_control = new BoxButtonControl();

    this.pack_end(this.box_boot_status, false, false, 3);
    this.pack_start(this.box_button_control, true, true, 3);
  }

  public void update_boot_status() {
    this.box_boot_status.update_boot_status();
  }

  public void set_surf_activated() {
    this.box_button_control.set_surf_activated();
  }

  public void set_surf_deactivated() {
    this.box_button_control.set_surf_deactivated();
  }
}


public class AnonSurfWindow: Gtk.ApplicationWindow {
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
    this.set_application_id("org.parrot.anonsurf.gui");
    /*
      Create a new window if app is activate
      Fix app created multiple windows
      Fix headless commands crashed because objects are not created
    */
    this.window = new AnonSurfWindow();
    this.activate.connect(activate_application);
  }

  private void create_new_window() {
    this.window = new AnonSurfWindow();
    this.window.show_all();
    this.refresh_gui();
  }

  public void activate_from_systray() {
    /*
      When left click on systray, this method is invoked
      Fix an other systray will be created, overlapping current systray when left click (bug happens when use activate application)
    */
    this.create_new_window();
  }

  private void activate_application() {
    /*
      When app is activated (left clicked from system tray menu), signal is_active is true
      Create new ApplicationWindow to fix error empty Application
      Fix the problem that systray only crashes when user runs anonsurf-gtk again
      Fix if user starts application, no window shows
    */
    this.create_new_window();

    /*
      There's a small delay if show_all is called after refresh gui.
      Use show_all first will avoid that delay
    */

    Timeout.add(400, refresh_gui); // Improve me: delay time should be better
    Gtk.main();
  }

  public void set_systray_connection(AnonSurfSystrayIcon systray) {
    this.sys_tray = systray;
  }

  private bool check_surf_running() {
    File file = File.new_for_path("/run/systemd/units/invocation:anonsurfd.service");
    return file.query_exists();
  }

  private bool check_tor_running() {
    File file = File.new_for_path("/run/systemd/units/invocation:tor.service");
    return file.query_exists();
  }

  private void set_tor_failed_to_run() {
    this.window.title_bar.set_tor_service_not_running();
  }

  private void set_surf_activated() {
    /*
    When AnonSurf is activated, it's necessary to check Tor's status as well.
    */
    if (!this.check_tor_running()) {
      this.set_tor_failed_to_run();
    } else {
      this.window.main_layout.set_surf_activated();
      this.window.title_bar.set_surf_activated();
    }
  }

  private void set_surf_deactivated() {
    this.window.main_layout.set_surf_deactivated();
    this.window.title_bar.set_surf_deactivated();
  }

  private bool refresh_gui() {
    /*
    Update system tray
    If visible:
    - Update title bar and main layout
    */
    bool is_surf_running = check_surf_running();
    if (is_surf_running) {
      this.sys_tray.set_surf_activated();
      if (!this.window.title_bar.get_visible()) {
        return true;
      }
      this.set_surf_activated();
      this.window.main_layout.box_boot_status.update_boot_status();
    } else {
      this.sys_tray.set_surf_deactivated();
      if (!this.window.title_bar.get_visible()) {
        return true;
      }
      this.set_surf_deactivated();
      this.window.main_layout.box_boot_status.update_boot_status();
    }
    return true;
  }
}


int main(string[] args) {
  Gtk.init(ref args);

  AnonSurfApp app = new AnonSurfApp();
  Notify.init("AnonSurf GTK");

  if (args.length == 1 && !app.window.is_active) {
    // FIXME activate multiple app still creates multiple systray layout
    AnonSurfSystrayIcon sys_tray = new AnonSurfSystrayIcon();

    sys_tray.set_app_connection(app);
    app.set_systray_connection(sys_tray);

    return app.run(args);
  } else {
    if (args[1] == "start" || args[1] == "stop") {
      app.window.main_layout.box_button_control.box_top_buttons.button_start.on_click_button();
    } else if (args[1] == "myip") {
      app.window.main_layout.box_button_control.box_top_buttons.button_my_ip.on_click_button();
    } else if (args[1] == "changeid") {
      app.window.main_layout.box_button_control.box_bottom_buttons.button_change_id.on_click_button.begin();
    } else {
      send_notification("Invalid option", "Option " + args[1] + " is invalid", NotifyLevel.Error);
      return 1;
    }
    return 0;
  }
}
