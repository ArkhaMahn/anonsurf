using Gtk;

const uchar BUTTON_SIZE_X = 105;
const uchar BUTTON_SIZE_Y = 65;
const uchar BUTTON_BOOT_SIZE_X = 60;
const uchar BUTTON_BOOT_SIZE_Y = 48;

public enum AnonSurfStatusErr {
  OK,
  TORNOTRUNNING,
  ANONSURFNOTRUNNING
}


public class ButtonStart: Button {
  public ThreadPool<ThreadStartAnonSurf> pool;
  private DialogAskKill dialog_ask_kill;

  public ButtonStart() {
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
    try {
      this.pool = new ThreadPool<ThreadStartAnonSurf>.with_owned_data((worker) => {
        worker.run();
      }, 1, false);
    } catch (GLib.Error error) {
      send_notification("Thread error", error.message, NotifyLevel.Error);
    }
  }

  public void on_click_button() {
    if (this.pool.get_num_threads() > 0) {
      return;
    }

    // Create a dialog to ask if user want to kill some processes and remove cache
    this.dialog_ask_kill = new DialogAskKill();
    dialog_ask_kill.start_dialog();

    // When dialog closes, start creating thread to either start or stop AnonSurf
    // Kill processes spawns before AnonSurf starts or stops
    // The kill proccess is handled by dialog, before kill signal is called
    this.dialog_ask_kill.destroy.connect(() => {
      try {
        if (this.get_label() == "Start") {
          this.pool.add(new ThreadStartAnonSurf(true));
        } else {
          this.pool.add(new ThreadStartAnonSurf(false));
        }
      } catch (ThreadError error) {
        send_notification("Error while stopping AnonSurf", error.message, NotifyLevel.Error);
      }
    });
  }

  public void check_thread_is_running() {
    if (this.pool.get_num_threads() > 0) {
      this.set_sensitive(false);
    } else {
      this.set_sensitive(true);
    }
  }

  public void set_surf_activated() {
    /*
      AnonSurf is activated. Change label to Stop
    */
    this.set_label("Stop");
  }

  public void set_surf_deactivated() {
    /*
      AnonSurf is deactivated. Change label to Start
    */
    this.set_label("Start");
  }
}


public class ButtonMyIP: Button {
  private ThreadPool<ThreadMyIP> pool;

  public ButtonMyIP() {
    this.set_label("My IP");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
    try {
      this.pool = new ThreadPool<ThreadMyIP>.with_owned_data((worker) => {
        worker.run();
      }, 1, false);
    } catch (GLib.Error error) {
      send_notification("Thread error", error.message, NotifyLevel.Error);
    }
  }

  public void on_click_button() {
    if (this.pool.get_num_threads() > 0) {
      return;
    }
    try {
      this.pool.add(new ThreadMyIP());
    } catch (ThreadError error) {
      send_notification("Thread error", error.message, NotifyLevel.Error);
    }
  }

  public void check_thread_is_running() {
    if (this.pool.get_num_threads() > 0) {
      this.set_sensitive(false);
    } else {
      this.set_sensitive(true);
    }
  }
}


public class ButtonChangeID: Button {
  private InetAddress address;
  private string cookie_hash;

  public ButtonChangeID() {
    this.set_label("Change ID"); // TODO improve button name
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
    this.address = new InetAddress.loopback(SocketFamily.IPV4);
  }

  private string get_path_cookie_from_torrc() {
    string cookie_path = "/var/run/tor/control.authcookie";
    File file = File.new_for_path("/etc/tor/torrc");

    try {
      FileInputStream @is = file.read();
      DataInputStream dis = new DataInputStream(@is);
      string line;

      while ((line = dis.read_line()) != null) {
        if (line.index_of("CookieAuthFile ") == 0) {
          // Line starts with CookieAuthFile
          cookie_path = line.split(" ")[1];
          @is.close();
          break;
        }
      }
      return cookie_path;

    } catch (Error error) {
      send_notification("Error reading Torrc", error.message, NotifyLevel.Error);
      return "";
    }
  }

  private void get_tor_cookie() {
    string cookie_path = this.get_path_cookie_from_torrc();
    File file = File.new_for_path(cookie_path);

    try {
      FileInputStream @is = file.read();
      DataInputStream dis = new DataInputStream(@is);
      this.cookie_hash = dis.read_line(); // Do we need to convert binary to hex?
    } catch (Error error) {
      send_notification("Error reading Tor's cookie auth", error.message, NotifyLevel.Error);
    }
  }

  public async void on_click_button() {
    // Get cookie when click change id
    // This would be much better if program gets tor cookie every time it starts
    // and remove when it closes
    this.get_tor_cookie();

    try {
      InetSocketAddress inet_address = new InetSocketAddress(this.address, 9051);
      SocketClient client = new SocketClient();
      SocketConnection conn = yield client.connect_async(inet_address);
      string message = "authenticate \"" + this.cookie_hash + "\"\n";
      DataInputStream input;

      yield conn.output_stream.write_async(message.data, Priority.DEFAULT);
      input = new DataInputStream(conn.input_stream);
      message = yield input.read_line_async();

      if (message == "250 OK\x0d") {
        send_notification("Authentication failed", message, NotifyLevel.Error);
        conn.close();
        return;
      }

      message = "signal newnym\n";
      yield conn.output_stream.write_async(message.data, Priority.DEFAULT);
      message = yield input.read_line_async();

      if (message == "250 OK\x0d") {
        send_notification("Tor's ID", "ID successfully changed", NotifyLevel.Ok); // TODO improve message
      } else {
        send_notification("Failed to change ID", message, NotifyLevel.Error);
      }

      message = "quit\n"; // Close connection
      yield conn.output_stream.write_async(message.data, Priority.DEFAULT);
      conn.close();

    } catch (Error error) {
      send_notification("Socket error", error.message, NotifyLevel.Error);
    }
  }

  public void set_surf_activated() {
    /*
      AnonSurf is activated. Allow clicking on this
    */
    this.set_sensitive(true);
  }


  public void set_surf_deactivated() {
    /*
      AnonSurf is deactivated. Disable clicking
    */
    this.set_sensitive(false);
  }
}


public class ButtonTorController: Button {
  public DialogTorController tor_status;

  public ButtonTorController() {
    this.set_label("Tor stat"); // TODO improve button name
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
  }

  public void on_click_button() {
    /*
      Create a dialog that shows a VTE with nyx
      - Check DialogTorController == null meant to prevent user creates multiple VTE (of nyx) windows
        However, it also prevents user starts VTE again (close and start) because Dialog is now not null anymore
      - Check !this.tor_status.get_visible() to show dialog from second time
    */
    // Prevent creating multiple dialogs when click on button
    if (this.tor_status == null || !this.tor_status.get_visible()) {
      this.tor_status = new DialogTorController();
      this.tor_status.show_dialog();
    }
  }

  public void set_surf_activated() {
    /*
      AnonSurf is activated. Allow clicking on this
    */
    this.set_sensitive(true);
  }


  public void set_surf_deactivated() {
    /*
      AnonSurf is deactivated. Disable clicking
    */
    this.set_sensitive(false);
  }
}


public class BoxTopButtons: Box {
  public ButtonStart button_start;
  public ButtonMyIP button_my_ip;

  public BoxTopButtons() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    this.button_start = new ButtonStart();
    this.button_my_ip = new ButtonMyIP();

    this.pack_start(button_start, false, true, 12);
    this.pack_start(button_my_ip, false, true, 10);
  }

  public void set_surf_activated() {
    this.button_start.set_surf_activated();
    this.button_start.check_thread_is_running();
    this.button_my_ip.check_thread_is_running();
  }

  public void set_surf_deactivated() {
    this.button_start.set_surf_deactivated();
    this.button_start.check_thread_is_running();
    this.button_my_ip.check_thread_is_running();
  }
}


public class BoxBottomButtons: Box {
  public ButtonChangeID button_change_id;
  public ButtonTorController button_tor_controller;

  public BoxBottomButtons() {
    GLib.Object(orientation: Orientation.HORIZONTAL);
    this.button_change_id = new ButtonChangeID();
    this.button_tor_controller = new ButtonTorController();

    this.pack_start(this.button_tor_controller, false, true, 12);
    this.pack_start(this.button_change_id, false, true, 10);
  }

  public void set_surf_activated() {
    this.button_change_id.set_surf_activated();
    this.button_tor_controller.set_surf_activated();
  }

  public void set_surf_deactivated() {
    this.button_change_id.set_surf_deactivated();
    this.button_tor_controller.set_surf_deactivated();
  }
}


public class BoxButtonControl: Box {
  public BoxBottomButtons box_bottom_buttons;
  public BoxTopButtons box_top_buttons;

  public BoxButtonControl() {
    GLib.Object(orientation: Orientation.VERTICAL);
    this.box_bottom_buttons = new BoxBottomButtons();
    this.box_top_buttons = new BoxTopButtons();

    this.pack_start(this.box_top_buttons, false, true, 3);
    this.pack_start(this.box_bottom_buttons, false, true, 3);
  }

  public void set_surf_activated() {
    this.box_top_buttons.set_surf_activated();
    this.box_bottom_buttons.set_surf_activated();
  }

  public void set_surf_deactivated() {
    this.box_top_buttons.set_surf_deactivated();
    this.box_bottom_buttons.set_surf_deactivated();
  }
}
