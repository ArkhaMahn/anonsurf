/*
Create a Terminal with VTE module to call the app Nyx which can control
Tor service
*/

using Vte;
using Gtk;

// FIXME: program nyx doesn't close when close dialog

public class NyxTerminal: Terminal {
  /*
  Create a terminal that spawns nyx to show Tor's status
   */
  private Dialog parent_dialog;

  public NyxTerminal(Dialog parent_dialog) {
    this.set_clear_background(true);
    this.set_cursor_blink_mode(ON);
    this.set_cursor_shape(UNDERLINE);
    this.parent_dialog = parent_dialog;
  }

  public void start_nyx_terminal() {
    this.spawn_async(
      DEFAULT, // https://valadoc.org/vte-2.91/Vte.PtyFlags.html
      null,
      {"/usr/bin/nyx", "--config", "/etc/anonsurf/nyxrc"}, // TODO remove nyx password
      {},
      CHILD_INHERITS_STDIN, // https://valadoc.org/glib-2.0/GLib.SpawnFlags.html
      null,
      -1,
      null, // callback?
      null
    );
    this.child_exited.connect(on_nyx_exit);
  }

  private void on_nyx_exit() {
    this.parent_dialog.hide();
  }

  // FIXME(anonsurf-gtk:367494): VTE-WARNING **: 18:17:01.056: (../src/vtepty.cc:810):void vte_pty_spawn_with_fds_async(VtePty*, const char*, const char* const*, const char* const*, const int*, int, const int*, int, GSpawnFlags, GSpawnChildSetupFunc, gpointer, GDestroyNotify, int, GCancellable*, GAsyncReadyCallback, gpointer): runtime check failed: ((spawn_flags & ignored_spawn_flags()) == 0)
}


public class DialogTorController: Dialog {
  private NyxTerminal vte_tor_controller;
  private Box content_area;
  private Box terminal_box;

  public DialogTorController() {
    this.set_title("Tor Status");
    this.set_resizable(false);
    this.set_icon_name("nyx-logo"); // FIXME must install this logo anyway

    vte_tor_controller = new NyxTerminal(this);
    terminal_box = new Box(Orientation.VERTICAL, 3);
    content_area = get_content_area() as Box;

    terminal_box.pack_start(vte_tor_controller, false, true, 3);
    content_area.pack_start(terminal_box, false, true, 3);
    // NOTICEME: when use show_all here, the dialog starts when program starts.
  }

  public void show_dialog() {
    this.vte_tor_controller.start_nyx_terminal(); // TODO double check to make sure nyx is killed when dialog closes
    this.show_all();
    this.run();
    this.hide();
  }
}
