using Gtk;


public class ImageAnonSurfStatus: Image {
  public void on_update_icon(bool is_surf_running) {
    if (is_surf_running) {
      // this.from_icon_name("security-high"); // GTK 4 syntax
      this.set_from_icon_name("security-high", DIALOG);
    } else {
      // this.from_icon_name("security-medium"); // GTK 4 syntax
      this.set_from_icon_name("security-medium", DIALOG);
    }
    // TODO show error too
  }
}


public class ButtonAbout: Button {
  public ButtonAbout() {
    this.set_image(new Image.from_icon_name("help-about", BUTTON));
    this.clicked.connect(on_click_about);
  }

  private void on_click_about() {
    var dialog_about = new DialogAboutAnonSurf();
    dialog_about.run();
    dialog_about.destroy();
  }
}


public class AnonSurfTitleBar: HeaderBar {
  private ImageAnonSurfStatus icon;
  private ButtonAbout button_about;

  public AnonSurfTitleBar() {
    icon = new ImageAnonSurfStatus();
    button_about = new ButtonAbout();

    this.set_show_close_button(true);
    this.set_decoration_layout(":close");
    this.set_title("AnonSurf");

    this.pack_start(icon);
    this.pack_end(button_about);
  }

  private void on_update_title(bool is_surf_running) {
    if (is_surf_running) {
      this.set_subtitle("Activated");
    } else {
      this.set_subtitle("Deactivated");
    }
    // TODO show error too
  }

  public void on_update_title_bar(bool is_surf_running) {
    if (this.get_visible()) {
      on_update_title(is_surf_running);
    }
    this.icon.on_update_icon(is_surf_running);
  }
}
