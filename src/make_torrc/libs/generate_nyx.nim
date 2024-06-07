#[
  Generate configurations for nyx
]#
import os
import .. / .. / dnstool / cli / print

const
  path_nyxrc = "/etc/anonsurf/nyxrc"
  config_nyx = "tor_chroot /var/run/tor/control"


proc maketorrc_nyxrc_checkconfig*() =
  #[
    Generate nyx's config file
    Since latest change (use tor's authentication instead of password everytime, this function should be called
    when file is different or doesn't exists only)
  ]#
  if not fileExists(path_nyxrc) or readFile(path_nyxrc) != config_nyx:
    try:
      writeFile(path_nyxrc, config_nyx)
      echo "New nyxrc created"
    except:
      print_error("Failed to create new nyxrc")
