using Gtk;


public class BoxMainLayout: Box {
  public BoxBootStatus box_boot_status;
  public BoxButtonControl box_button_control;

  public BoxMainLayout() {
    GLib.Object(orientation: Orientation.VERTICAL);
    this.box_boot_status = new BoxBootStatus();
    this.box_button_control = new BoxButtonControl();

    this.pack_start(this.box_button_control, false, false, 3);
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
    this.set_icon_name("anonsurf");
    this.set_resizable(false);
    this.add(this.main_layout);
  }
}


public class AnonSurfApp: GLib.Application {
  public AnonSurfWindow window;
  private AnonSurfSystrayIcon sys_tray;

  public AnonSurfApp() {
    /*
      Create a new Application
      Startup signal will create an unique application, which fixed:
        1. multiple windows are created
        2. multiple system tray icons are created
    */
    this.set_application_id("org.parrot.anonsurf.gui");
    this.startup.connect(create_window);
    // this.activate.connect(restore_window);
  }

  private void create_window() {
    /*
      Create a new window when Application is created
    */
    this.window = new AnonSurfWindow();
    this.window.show_all();
    this.create_system_tray();
    this.refresh_gui();

    this.activate.connect(restore_window);

    /*
      There's a small delay if show_all is called after refresh gui.
      Use show_all first will avoid that delay
      ImproveMe: delay time should be better
    */
    Timeout.add(400, refresh_gui);
    Gtk.main();
  }

  public void create_headless_mode() {
    this.window = new AnonSurfWindow();

    /*
      Only update status of the controller buttons. So when user launches
      program from menu (quick start / stop) -> program automatically decides start / stop service
    */
    if (check_surf_running()) {
      this.set_surf_activated();
    } else {
      this.set_surf_deactivated();
    }
  }

  private void create_system_tray() {
    this.sys_tray = new AnonSurfSystrayIcon();

    this.sys_tray.set_app_connection(this);
    this.set_systray_connection(this.sys_tray);
  }

  public void restore_window() {
    /*
      GTK application gets "activate" signal when
        1. User do left click on system tray icon
        2. When window is closed, and user runs anonsurf-gtk
      This function restores the main window
      Create a new AnonSurfWindow object fixes:
        - Starts anonsurf-gtk when window was closed causes a crash, or show empty dialog
        - Create an other window when user left click on system tray icon
    */
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
    File file = File.new_for_path("/run/systemd/units/invocation:anonsurfd.service");
    return file.query_exists();
  }

  private bool check_tor_running() {
    File file = File.new_for_path("/run/systemd/units/invocation:tor.service");
    return file.query_exists();
  }

  private void set_tor_failed_to_run() {
    this.window.title_bar.set_tor_service_not_running();
    // When surf is running, but Tor doesn't run, button Stop needs to be there anyway.
    // Fix problem GUI basically in a hang mode
    this.window.main_layout.box_button_control.box_top_buttons.button_start.set_surf_activated();
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

  if (args.length == 1) {
    return app.run(args);
  } else {
    app.create_headless_mode();

    if (args[1] == "start" || args[1] == "stop") { // This is working without notification
      app.window.main_layout.box_button_control.box_top_buttons.button_start.on_click_button();
    } else if (args[1] == "myip") { // FIXME: this doesn't work (both notifi & function)
      app.window.main_layout.box_button_control.box_top_buttons.button_my_ip.on_click_button();
    } else if (args[1] == "changeid") { // Likely it's working without notification
      app.window.main_layout.box_button_control.box_bottom_buttons.button_change_id.on_click_button.begin();
    } else {
      send_notification("Invalid option", "Option " + args[1] + " is invalid", NotifyLevel.Error);
      return 1;
    }
    return 0;
  }
}
