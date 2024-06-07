import os
import libs / [generate_torrc, generate_nyx]


proc main() =
  #[
    Handle options
    if no flag, or flag == new-torrc -> create new torrc
    if flag == restore -> restore default config
  ]#
  if paramCount() == 0:
    #[
      In old version, we are having the same logic
      No need to change bash file for this
    ]#
    maketorrc_generate_torrc_config()
    maketorrc_nyxrc_checkconfig()
  elif paramCount() > 1:
    return
  else:
    if paramStr(1) == "restore":
      makettorc_restore_torrc_backup()
    elif paramStr(1) == "new-torrc":
      maketorrc_generate_torrc_config()

main()
