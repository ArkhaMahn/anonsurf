using Gtk;


public class SystrayMenuStart: Gtk.MenuItem {
  public SystrayMenuStart() {
    this.set_label("Start");
    this.activate.connect(on_click_main_action);
  }

  private void on_click_main_action() {
    if (this.get_label() == "Start") {
      print("Start\n");
    }
    else {
      print("Stop\n");
    }
  }

  public void on_update_status(bool is_surf_running) {
    if (is_surf_running) {
      this.set_label("Stop");
    } else {
      this.set_label("Start");
    }
  }
}


public class SystrayMenuStatus: Gtk.MenuItem {
  public SystrayMenuStatus() {
    this.set_label("Status");
    this.activate.connect(on_click_check_status);
  }

  private void on_click_check_status() {
    print("Clicked check status");
  }
}


public class SystrayMenuCheckIP: Gtk.MenuItem {
  public SystrayMenuCheckIP() {
    this.set_label("Check IP");
    this.activate.connect(on_click_check_ip);
  }

  private void on_click_check_ip() {
    print("Clicked check ip\n");
  }
}


public class SystrayMenuQuit: Gtk.MenuItem {
  public SystrayMenuQuit() {
    this.set_label("Quit");
    this.activate.connect(Gtk.main_quit);
  }
}


public class AnonSurfSystrayMenu: Gtk.Menu {
  public SystrayMenuStart menu_start;
  public SystrayMenuStatus menu_status;
  public SystrayMenuCheckIP menu_check_ip;
  public SystrayMenuQuit menu_quit;

  public AnonSurfSystrayMenu() {
    this.menu_start = new SystrayMenuStart();
    this.menu_status = new SystrayMenuStatus();
    this.menu_check_ip = new SystrayMenuCheckIP();
    this.menu_quit = new SystrayMenuQuit();

    // FIXME Incompatible pointer when append
    this.append(this.menu_start);
    this.append(this.menu_status);
    this.append(this.menu_check_ip);
    this.append(this.menu_quit);
    this.show_all();
  }
}


public class AnonSurfSystrayIcon: StatusIcon {
  private AnonSurfSystrayMenu menu;
  private AnonSurfApp app;

  public AnonSurfSystrayIcon() {
    this.menu = new AnonSurfSystrayMenu();
    this.set_from_icon_name("anonsurf");

    this.activate.connect(on_left_click_systray);
    this.popup_menu.connect(on_right_click_systray);
  }

  private void on_right_click_systray(uint button, uint time) {
    this.menu.popup_at_pointer();
  }

  public void set_app_connection(AnonSurfApp app) {
    this.app = app;
  }

  private void on_left_click_systray() {
    this.app.activate();
  }

  public void on_update_systray(bool is_surf_running) {
    this.menu.menu_start.on_update_status(is_surf_running);
  }
}
