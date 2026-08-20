/*
  Show AboutDialog on Title bar
 */

using Gtk;


public class DialogAboutAnonSurf: AboutDialog {
  public DialogAboutAnonSurf() {
    const string authors[] = {
      "Lorenzo \"Palinuro\" Faletra",
      "Nong Hoang \"DmKnght\" Tu",
      "Lisetta \"Sheireen\" Ferrero",
      "Francesco \"Mibofra\" Bonanno",
      "Manuel \"Serverket\" Hernandez",
    };
    const string artists[] = {
      "Federica \"marafed\" Marasà",
      "David \"mcder3\" Linares",
    };

    this.set_icon_name("anonsurf");
    this.set_logo_icon_name("anonsurf");
    this.set_program_name("AnonSurf");
    this.set_version("6.0.0");
    this.set_artists(artists);
    this.set_authors(authors);
    this.set_comments("Anonymous Toolkit of ParrotOS");
    this.set_copyright("Copyright © 2013 - 2024 Lorenzo \"Palinuro\" Faletra\nCopyright © 2022 Parrot Security CIC");
    this.set_license_type(GPL_3_0);
    this.set_website("https://gitlab.com/parrotsec/packages/anonsurf/");
    this.set_website_label("Gitlab Source");
  }
}
