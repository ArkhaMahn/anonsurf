using Vte;
using Gtk;


public class NyxTerminal: Terminal {
  /*
  Create a terminal that spawns nyx to show Tor's status
   */
  // public NyxTerminal(Dialog parent_dialog) {
  public NyxTerminal() {
    // FIXME for some reason, nyx doesn't take the config when program starts (or service). Wrong spawn_async call?
    this.set_clear_background(true);
    this.set_cursor_blink_mode(ON);
    this.set_cursor_shape(UNDERLINE);
    this.spawn_async(
      DEFAULT, // https://valadoc.org/vte-2.91/Vte.PtyFlags.html
      null,
      {"/usr/bin/nyx", "--config", "/etc/anonsurf/nyxrc"}, // TODO remove nyx password
      {},
      DO_NOT_REAP_CHILD, // https://valadoc.org/glib-2.0/GLib.SpawnFlags.html
      null,
      -1,
      null, // callback?
      null
    );
    // https://valadoc.org/vte-2.91/Vte.Terminal.child_exited.html
    // this.connect("child_exited", on_nyx_exit, parent_dialog);  // FIXME  g_object_connect: invalid signal spec "child-exited"
  }

  // private void on_nyx_exit(int signal, Dialog parent_dialog) {
  //   parent_dialog.destroy();
  // }
  // FIXME(anonsurf-gtk:367494): VTE-WARNING **: 18:17:01.056: (../src/vtepty.cc:810):void vte_pty_spawn_with_fds_async(VtePty*, const char*, const char* const*, const char* const*, const int*, int, const int*, int, GSpawnFlags, GSpawnChildSetupFunc, gpointer, GDestroyNotify, int, GCancellable*, GAsyncReadyCallback, gpointer): runtime check failed: ((spawn_flags & ignored_spawn_flags()) == 0)
}


public class StatusDialog: Dialog {
  private Terminal term_status;
  private Box content_area;
  private Box terminal_box;

  public StatusDialog() {
    this.set_title("Tor Status");
    this.set_resizable(false);
    this.set_icon_name("nyx-logo"); // FIXME must install this logo anyway
    this.destroy.connect(() => {Gtk.main_quit();});


    // term_status = new NyxTerminal(this);
    term_status = new NyxTerminal();
    terminal_box = new Box(Orientation.VERTICAL, 3);
    content_area = get_content_area() as Box;

    terminal_box.pack_start(term_status, false, true, 3);
    content_area.pack_start(terminal_box, false, true, 3);
    // NOTICEME: when use show_all here, the dialog starts when program starts.
  }

  public void show_dialog() {
    this.show_all();
    this.run();
    this.hide();
  }
}
