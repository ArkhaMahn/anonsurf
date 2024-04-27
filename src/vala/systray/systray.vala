using Gtk;


public class SystrayItemStart: Gtk.MenuItem {
  private AnonSurfApp app;

  public SystrayItemStart() {
    this.set_label("Start");
    this.activate.connect(on_click_main_action);
  }

  private void on_click_main_action() {
    this.app.window.main_layout.boxButton.boxTop.btnStart.on_click_button();
  }

  public void on_update_status(bool is_surf_running) {
    if (is_surf_running) {
      this.set_label("Stop");
    } else {
      this.set_label("Start");
    }
  }

  public void connect_anonsurf_button(AnonSurfApp app) {
    this.app = app;
  }
}


public class SystrayItemStatus: Gtk.MenuItem {
  // TODO prevent this menu spawns multiple dialogs
  private AnonSurfApp app;

  public SystrayItemStatus() {
    this.set_label("Status");
    this.activate.connect(on_click_check_status);
  }

  private void on_click_check_status() {
    this.app.window.main_layout.boxButton.boxBottom.btnStatus.on_click_button();
  }

  public void connect_anonsurf_button(AnonSurfApp app) {
    this.app = app;
  }
}


public class SystrayItemChangeId: Gtk.MenuItem {
  private AnonSurfApp app;

  public SystrayItemChangeId() {
    this.set_label("Change Identity");
    this.activate.connect(on_click_change_id);
  }

  private void on_click_change_id() {
    this.app.window.main_layout.boxButton.boxBottom.btnChangeID.on_click_button();
  }

  public void connect_anonsurf_button(AnonSurfApp app) {
    this.app = app;
  }
}


public class SystrayItemCheckIP: Gtk.MenuItem {
  private AnonSurfApp app;

  public SystrayItemCheckIP() {
    this.set_label("Check IP");
    this.activate.connect(on_click_check_status);
  }

  private void on_click_check_status() {
    this.app.window.main_layout.boxButton.boxTop.btnMyIP.on_click_button();
  }

  public void connect_anonsurf_button(AnonSurfApp app) {
    this.app = app;
  }
}


public class SystrayItemQuit: Gtk.MenuItem {
  public SystrayItemQuit() {
    this.set_label("Quit");
    this.activate.connect(Gtk.main_quit);
  }
}


public class AnonSurfSystrayMenu: Gtk.Menu {
  public SystrayItemStart menu_item_start;
  public SystrayItemStatus menu_item_status;
  public SystrayItemChangeId menu_item_change_id;
  public SystrayItemCheckIP menu_item_check_ip;
  public SystrayItemQuit menu_item_quit;

  public AnonSurfSystrayMenu() {
    this.create_menu_items();
    this.show_menu_items();
    // Refresher will be called by main Application so no need to call it here
  }

  private void create_menu_items() {
    this.menu_item_start = new SystrayItemStart();
    this.menu_item_status = new SystrayItemStatus();
    this.menu_item_change_id = new SystrayItemChangeId();
    this.menu_item_check_ip = new SystrayItemCheckIP();
    this.menu_item_quit = new SystrayItemQuit();

    // FIXME Vala warns incompatible pointer when append
    this.append(this.menu_item_start);
    this.append(this.menu_item_status);
    this.append(this.menu_item_change_id);
    this.append(this.menu_item_check_ip);
    this.append(this.menu_item_quit);
  }

  private void show_menu_items() {
    /*
    There are some items should be shown without checking AnonSurf's status
    */
    this.menu_item_start.show();
    this.menu_item_check_ip.show();
    this.menu_item_quit.show();
  }

  public void on_update_menu_list(bool is_surf_running) {
    if (is_surf_running) {
      this.menu_item_status.show();
      this.menu_item_change_id.show();
    } else {
      this.menu_item_status.hide();
      this.menu_item_change_id.hide();
    }
  }
}


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

    this.menu.menu_item_start.connect_anonsurf_button(app);
    this.menu.menu_item_status.connect_anonsurf_button(app);
    this.menu.menu_item_change_id.connect_anonsurf_button(app);
    this.menu.menu_item_check_ip.connect_anonsurf_button(app);
  }

  public void on_update_systray(bool is_surf_running) {
    this.menu.menu_item_start.on_update_status(is_surf_running);
    this.menu.on_update_menu_list(is_surf_running);

    if (is_surf_running) {
      this.set_from_icon_name("security-high");
    } else {
      this.set_from_icon_name("anonsurf");
    }
  }
}
