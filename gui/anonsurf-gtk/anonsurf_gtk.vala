using Gtk;

public bool systemd_unit_matches(string action, string unit) {
  try {
    Subprocess process = new Subprocess.newv({
      "/usr/bin/systemctl",
      "--quiet",
      action,
      unit
    }, STDOUT_SILENCE | STDERR_SILENCE);
    process.wait();
    return process.get_successful();
  } catch (GLib.Error error) {
    return false;
  }
}

public class BoxMainLayout: Box {
  public BoxBootStatus box_boot_status;
  public BoxButtonControl box_button_control;

  public BoxMainLayout() {
    GLib.Object(orientation: Orientation.VERTICAL);
    this.box_boot_status = new BoxBootStatus();
    this.box_button_control = new BoxButtonControl();

    this.pack_start(this.box_button_control, false, false, 6);
    this.pack_end(this.box_boot_status, false, false, 3);
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
    this.set_resizable(false);

    try {
      this.set_icon_from_file("/usr/share/icons/anonsurf.png");
    } catch {
      this.set_icon_name("anonsurf");
    }

    this.add(this.main_layout);
  }
}


public class AnonSurfApp: GLib.Application {
  public AnonSurfWindow window;
  private AnonSurfSystrayIcon sys_tray;

  public AnonSurfApp() {
    this.set_application_id("org.parrot.anonsurf.gui");
    this.startup.connect(create_window);
  }

  private void create_window() {
    this.window = new AnonSurfWindow();
    this.window.show_all();
    this.create_system_tray();
    this.refresh_gui();

    this.activate.connect(restore_window);
    Timeout.add_seconds(2, refresh_gui);
    Gtk.main();
  }

  private void create_system_tray() {
    this.sys_tray = new AnonSurfSystrayIcon();

    this.sys_tray.set_app_connection(this);
    this.set_systray_connection(this.sys_tray);
  }

  public void restore_window() {
    if (!this.window.get_visible()) {
      this.window = new AnonSurfWindow();
      this.window.show_all();
      this.refresh_gui();
    }
  }

  public void set_systray_connection(AnonSurfSystrayIcon systray) {
    this.sys_tray = systray;
  }

  private bool check_surf_running() {
    return systemd_unit_matches("is-active", "anonsurfd.service");
  }

  private bool check_tor_running() {
    return systemd_unit_matches("is-active", "tor@default.service");
  }

  private void set_tor_failed_to_run() {
    this.window.title_bar.set_tor_service_not_running();
    this.window.main_layout.box_button_control.box_top_buttons.button_start.set_surf_activated();
    this.window.main_layout.box_button_control.box_top_buttons.button_start.set_sensitive(true);
  }

  private void set_surf_activated() {
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
    bool is_surf_running = check_surf_running();
    if (is_surf_running) {
      this.sys_tray.set_surf_activated();
      if (!this.window.get_visible()) {
        return true;
      }
      this.set_surf_activated();
      this.window.main_layout.box_boot_status.update_boot_status();
    } else {
      this.sys_tray.set_surf_deactivated();
      if (!this.window.get_visible()) {
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

  Notify.init("AnonSurf GTK");
  AnonSurfApp app = new AnonSurfApp();
  return app.run(args);
}
