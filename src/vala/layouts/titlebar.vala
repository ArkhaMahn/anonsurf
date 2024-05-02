using Gtk;


public class ImageAnonSurfStatus: Image {
  public void set_surf_activated() {
    this.set_from_icon_name("security-high", DIALOG);
  }

  public void set_surf_deactivated() {
    this.set_from_icon_name("security-medium", DIALOG);
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

    this.set_title("AnonSurf");
    this.set_show_close_button(true);
    this.set_decoration_layout(":close");

    this.pack_start(image_anonsurf_status);
    this.pack_end(button_show_dialog_about);
  }

  public void set_surf_activated() {
    this.set_subtitle("Activated");
    this.image_anonsurf_status.set_surf_activated();
  }

  public void set_surf_deactivated() {
    this.set_subtitle("Deactivated");
    this.image_anonsurf_status.set_surf_deactivated();
  }
}
