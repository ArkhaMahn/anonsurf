using Gtk;


public class ButtonAbout: Button {
  public ButtonAbout() {
    this.set_image(new Image.from_icon_name("help-about", BUTTON));
    this.clicked.connect(on_click_about);
  }

  private void on_click_about() {
    var dialog_about = new AnonSurfDialogAbout();
    dialog_about.run();
    dialog_about.destroy();
  }
}


public class AnonSurfTitleBar: HeaderBar {
  public AnonSurfTitleBar() {
    this.set_show_close_button(true);
    this.set_title("AnonSurf"); // TODO remove this
    this.pack_end(new ButtonAbout());
    this.set_decoration_layout(":close");
  }

  public void on_update_title(bool is_surf_running) {
    if (is_surf_running) {
      this.set_title("AnonSurf - Activated");
    } else {
      this.set_title("AnonSurf - Deactivated");
    }
    // TODO show error too
  }
}
