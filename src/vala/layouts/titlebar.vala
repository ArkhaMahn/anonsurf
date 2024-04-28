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


public class ButtonShowAboutAnonSurf: Button {
  public ButtonShowAboutAnonSurf() {
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
  private ImageAnonSurfStatus image_anonsurf_status;
  private ButtonShowAboutAnonSurf button_show_dialog_about;

  public AnonSurfTitleBar() {
    image_anonsurf_status = new ImageAnonSurfStatus();
    button_show_dialog_about = new ButtonShowAboutAnonSurf();

    this.set_show_close_button(true);
    this.set_decoration_layout(":close");
    this.set_title("AnonSurf");

    this.pack_start(image_anonsurf_status);
    this.pack_end(button_show_dialog_about);
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
