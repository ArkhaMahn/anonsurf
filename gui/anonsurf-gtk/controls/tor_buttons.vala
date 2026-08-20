using Gtk;

public class ButtonChangeID: Button {
  private InetAddress address;
  private string cookie_hash;

  public ButtonChangeID() {
    this.set_label("Change ID");
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

  private bool get_tor_cookie() {
    string cookie_path = this.get_path_cookie_from_torrc();
    File file = File.new_for_path(cookie_path);

    try {
      StringBuilder hex_str = new StringBuilder();
      FileInputStream @is = file.read();
      DataInputStream dis = new DataInputStream(@is);
      string cookie_binary = dis.read_line();

      for (char i = 0; i < cookie_binary.length; i ++) {
        hex_str.append_printf("%02x", (uchar)cookie_binary[i]);
      }

      this.cookie_hash = hex_str.str;
      return true;
    } catch (Error error) {
      send_notification("Error reading Tor's cookie auth", error.message, NotifyLevel.Error);
      return false;
    }
  }

  public async void on_click_button() {
    if (!this.get_tor_cookie()) {
      return;
    }

    try {
      InetSocketAddress inet_address = new InetSocketAddress(this.address, 9051);
      SocketClient client = new SocketClient();
      SocketConnection conn = yield client.connect_async(inet_address);
      string message = "authenticate " + this.cookie_hash + "\n";
      DataInputStream input;

      yield conn.output_stream.write_async(message.data, Priority.DEFAULT);
      input = new DataInputStream(conn.input_stream);
      message = yield input.read_line_async();

      if (message != "250 OK\x0d") {
        send_notification("Change ID failed", message, NotifyLevel.Error);
        conn.close();
        return;
      }

      message = "signal newnym\n";
      yield conn.output_stream.write_async(message.data, Priority.DEFAULT);
      message = yield input.read_line_async();

      if (message == "250 OK\x0d") {
        send_notification("Change ID succeeded", "ID successfully changed", NotifyLevel.Ok);
      } else {
        send_notification("Change ID failed", message, NotifyLevel.Error);
      }

      message = "quit\n";
      yield conn.output_stream.write_async(message.data, Priority.DEFAULT);
      conn.close();
    } catch (Error error) {
      send_notification("Change ID failed", error.message, NotifyLevel.Error);
    }
  }

  public void set_surf_activated() {
    this.set_sensitive(true);
  }

  public void set_surf_deactivated() {
    this.set_sensitive(false);
  }
}

public class ButtonTorController: Button {
  public DialogTorController tor_status;

  public ButtonTorController() {
    this.set_label("Tor status");
    this.set_size_request(BUTTON_SIZE_X, BUTTON_SIZE_Y);
    this.clicked.connect(on_click_button);
  }

  public void on_click_button() {
    if (this.tor_status == null || !this.tor_status.get_visible()) {
      this.tor_status = new DialogTorController();
      this.tor_status.show_dialog();
    }
  }

  public void set_surf_activated() {
    this.set_sensitive(true);
  }

  public void set_surf_deactivated() {
    this.set_sensitive(false);
  }
}
