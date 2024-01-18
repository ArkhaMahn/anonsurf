import os
import libs / [cli_utils, boot_utils, myip, tor_utils]



#[
  Check user's options
]#
proc checkOptions() =
  if paramCount() != 1:
    cli_banner_about()
    cli_banner_help()
  else:
    case paramStr(1)
    of "start":
      cli_anonsurf_start()
    of "stop":
      cli_anonsurf_stop()
    of "status":
      cli_anonsurf_tor_status()
    of "enable-boot":
      cli_anonsurf_enable_boot()
    of "disable-boot":
      cli_anonsurf_disable_boot()
    of "status-boot":
      cli_anonsurf_check_boot()
    of "changeid":
      cli_anonsurf_change_id()
    of "myip":
      cli_check_current_ip()
    of "help":
      cli_banner_about()
      cli_banner_help()
    else:
      cli_send_msg(Err, "Invalid option", "Invalid option " & paramStr(1))

checkOptions()
