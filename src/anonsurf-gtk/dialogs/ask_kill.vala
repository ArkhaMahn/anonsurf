using Gtk;


class ButtonKillYes: Button {
  public ButtonKillYes() {
    this.set_label("Kill");
    this.clicked.connect(on_confirm_kill);
  }

  private void on_confirm_kill() {
    try {
      Subprocess action_start_anonsurf = new Subprocess.newv({"/usr/lib/anonsurf/safekill"}, NONE);
      action_start_anonsurf.wait();
      send_notification("Clear cache completed", "Killed processes and clear cache", NotifyLevel.Ok);
    } catch (GLib.Error error) {
      send_notification("Failed to do clear cache", error.message, NotifyLevel.Error);
    }
  }
}


class ButtonKillNo: Button {
  Dialog parent_dialog;

  public ButtonKillNo() {
    this.set_label("Cancel");
    this.clicked.connect(on_no_kill);
  }

  private void on_no_kill() {
    this.parent_dialog.destroy();
  }

  public void set_dialog_relationship(Dialog d) {
    this.parent_dialog = d;
  }
}


class DialogAskKill: Dialog {
  private Box selection_box;
  private Box content_area;
  private ButtonKillYes yes_button;
  private ButtonKillNo no_button;

  public DialogAskKill() {
    this.set_title("Kill processes and clean cache");
    this.set_resizable(false);

    this.make_dialog_contents();
    // TODO add some labels to tell user what the f is this
    // TODO better UX, add icon
  }

  private void make_dialog_contents() {
    this.yes_button = new ButtonKillYes();
    this.no_button = new ButtonKillNo();

    this.no_button.set_dialog_relationship(this);

    this.content_area = get_content_area() as Box;
    this.selection_box = new Box(Orientation.HORIZONTAL, 3);

    this.selection_box.pack_start(this.yes_button);
    this.selection_box.pack_start(this.no_button);

    this.content_area.pack_end(selection_box, false, true, 3);
  }

  public void start_dialog() {
    this.show_all();
  }
}