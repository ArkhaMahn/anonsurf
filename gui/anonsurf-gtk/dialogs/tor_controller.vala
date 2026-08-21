using Vte;
using Gtk;


public class NyxTerminal: Terminal {
  private Dialog parent_dialog;

  public NyxTerminal(Dialog parent_dialog) {
    this.set_clear_background(true);
    this.set_cursor_blink_mode(ON);
    this.set_cursor_shape(UNDERLINE);
    this.parent_dialog = parent_dialog;
  }

  public void start_nyx_terminal() {
    this.spawn_async(
      DEFAULT,
      null,
      {"/usr/bin/pkexec", "/usr/lib/anonsurf/anondaemon", "monitor"},
      {},
      SEARCH_PATH,
      null,
      -1,
      null,
      null
    );
    this.child_exited.connect(on_nyx_exit);
  }

  private void on_nyx_exit() {
    this.parent_dialog.destroy();
  }
}


public class DialogTorController: Dialog {
  private NyxTerminal vte_tor_controller;
  private Box content_area;

  public DialogTorController() {
    this.vte_tor_controller = new NyxTerminal(this);
    this.set_title("Tor Status");
    this.set_resizable(false);
    try {
      this.set_icon_from_file("/usr/share/icons/nyx-logo.png");
    } catch {
      this.set_icon_name("nyx-logo");
    }
  }

  public void show_dialog() {
    content_area = get_content_area() as Box;
    content_area.pack_start(this.vte_tor_controller, false, true, 3);
    this.vte_tor_controller.start_nyx_terminal();
    this.show_all();
  }
}
