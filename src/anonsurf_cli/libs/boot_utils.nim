import os
import print_utils


proc is_anonsurf_enabled_systemd(): bool =
  if fileExists("/etc/systemd/system/multi-user.target.wants/anonsurfd.service"):
    return true
  elif fileExists("/etc/systemd/system/default.target.wants/anonsurfd.service"):
    return true
  return false


#[
  Check if anonsurf is enabled
]#
proc cli_anonsurf_check_boot*() =
  if is_anonsurf_enabled_systemd():
    cli_send_msg(Ok, "AnonSurf is enabled", "AnonSurf is enabled and can start with system")
  else:
    cli_send_msg(Warn, "AnonSurf is enabled", "AnonSurf is disabled. Your connection is not under Tor network")


#[
  Use systemctl to enable anonsurf at boot
]#
proc do_enable_boot(): int =
  return execShellCmd("sudo /usr/bin/systemctl enable anonsurfd")


#[
  Enable anonsurf at boot
]#
proc cli_anonsurf_enable_boot*() =
  if is_anonsurf_enabled_systemd():
    cli_send_msg(Err, "AnonSurf is already enabled", "Anonsurf is enabled. No need to enable it")
  else:
    if do_enable_boot() == 0:
      cli_send_msg(Ok, "AnonSurf is enabled", "Your traffic is under Tor network when system starts")
    else:
      cli_send_msg(Err, "Failed to enable AnonSurf", "An error happened when enable AnonSurf at boot")


#[
  Use systemctl to disable anonsurf at boot
]#
proc do_disable_boot(): int =
  return execShellCmd("sudo /usr/bin/systemctl disable anonsurfd")


#[
  Disable anonsurf at boot
]#
proc cli_anonsurf_disable_boot*() =
  if not is_anonsurf_enabled_systemd():
    cli_send_msg(Err, "AnonSurf is disabled", "AnonSurf is disabled. No need to enable it")
  else:
    if do_disable_boot() == 0:
      cli_send_msg(Ok, "AnonSurf is disabled", "Your traffic is no longer anonimimi")
    else:
      cli_send_msg(Err, "Failed to disable AnonSurf", "An error happened when disable AnonSurf")
