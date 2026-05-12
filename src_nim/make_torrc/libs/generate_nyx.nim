#[
  Generate configurations for nyx
]#
import os
import strutils
import .. / .. / dnstool / cli / print

const
  path_nyxrc = "/etc/anonsurf/nyxrc"
  config_nyx = staticRead("../../../configs/nyxrc")


proc maketorrc_nyxrc_checkconfig*() =
  #[
    Generate nyx's config file
    Since latest change (use tor's authentication instead of password everytime, this function should be called
    when expected config not in the file)
  ]#
  if not fileExists(path_nyxrc) or (config_nyx notin readFile(path_nyxrc)):
    try:
      writeFile(path_nyxrc, config_nyx)
      echo "New nyxrc created"
    except:
      print_error("Failed to create new nyxrc")
