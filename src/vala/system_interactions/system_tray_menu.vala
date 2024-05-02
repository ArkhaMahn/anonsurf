/*
Create a menu when user right click on system tray's icon
The items in the menu will be updated (hide or show)
depend on AnonSurf's status
TODO instead of only AnonSurf, the update should consider any error of Tor too
*/

using Gtk;


public class MenuItemStart: Gtk.MenuItem {
  private AnonSurfApp app;

  public MenuItemStart() {
    this.activate.connect(on_click_main_action);
  }

  private void on_click_main_action() {
    this.app.window.main_layout.box_button_control.box_top_buttons.button_start.on_click_button();
  }

  public void update_label(bool is_surf_running) {
    if (is_surf_running) {
      this.set_label("Stop");
    } else {
      this.set_label("Start");
    }
  }

  public void connect_anonsurf_button(AnonSurfApp app) {
    this.app = app;
  }

  public void set_surf_activated() {
    this.set_label("Stop");
  }

  public void set_surf_deactivated() {
    this.set_label("Start");
  }
}


public class MenuItemTorController: Gtk.MenuItem {
  // TODO prevent this menu spawns multiple dialogs
  private AnonSurfApp app;

  public MenuItemTorController() {
    this.set_label("Tor Controller");
    this.activate.connect(on_click_check_status);
  }

  private void on_click_check_status() {
    this.app.window.main_layout.box_button_control.box_bottom_buttons.button_tor_controller.on_click_button();
  }

  public void connect_anonsurf_button(AnonSurfApp app) {
    this.app = app;
  }
}


public class MenuItemItemChangeId: Gtk.MenuItem {
  private AnonSurfApp app;

  public MenuItemItemChangeId() {
    this.set_label("Change Identity");
    this.activate.connect(on_click_change_id);
  }

  private void on_click_change_id() {
    this.app.window.main_layout.box_button_control.box_bottom_buttons.button_change_id.on_click_button();
  }

  public void connect_anonsurf_button(AnonSurfApp app) {
    this.app = app;
  }
}


public class MenuItemCheckIP: Gtk.MenuItem {
  private AnonSurfApp app;

  public MenuItemCheckIP() {
    this.set_label("Check IP");
    this.activate.connect(on_click_check_status);
  }

  private void on_click_check_status() {
    this.app.window.main_layout.box_button_control.box_top_buttons.button_my_ip.on_click_button();
  }

  public void connect_anonsurf_button(AnonSurfApp app) {
    this.app = app;
  }
}


public class MenuItemQuit: Gtk.MenuItem {
  public MenuItemQuit() {
    this.set_label("Quit");
    this.activate.connect(Gtk.main_quit);
  }
}


public class AnonSurfSystrayMenu: Gtk.Menu {
  public MenuItemStart menu_item_start;
  public MenuItemTorController menu_item_tor_controller;
  public MenuItemItemChangeId menu_item_change_id;
  public MenuItemCheckIP menu_item_check_ip;
  public MenuItemQuit menu_item_quit;

  public AnonSurfSystrayMenu() {
    this.create_menu_items();
    this.show_menu_items();
    // Refresher will be called by main Application so no need to call it here
  }

  private void create_menu_items() {
    this.menu_item_start = new MenuItemStart();
    this.menu_item_tor_controller = new MenuItemTorController();
    this.menu_item_change_id = new MenuItemItemChangeId();
    this.menu_item_check_ip = new MenuItemCheckIP();
    this.menu_item_quit = new MenuItemQuit();

    // FIXME Vala warns incompatible pointer when append
    this.append(this.menu_item_start);
    this.append(this.menu_item_tor_controller);
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

  public void update_list_items(bool is_surf_running) {
    if (is_surf_running) {
      this.menu_item_tor_controller.show();
      this.menu_item_change_id.show();
    } else {
      this.menu_item_tor_controller.hide();
      this.menu_item_change_id.hide();
    }
  }

  public void set_surf_activated() {
    this.menu_item_tor_controller.show();
    this.menu_item_change_id.show();
    this.menu_item_start.set_surf_activated();
  }

  public void set_surf_deactivated() {
    this.menu_item_tor_controller.hide();
    this.menu_item_change_id.hide();
    this.menu_item_start.set_surf_deactivated();
  }
}
