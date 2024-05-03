import os
import posix
import strscans
import net
import strutils
import cli_utils


# TODO kill apps

#[
  Check if anonsurfd is running using a symlink in /run/
]#
proc is_anonsurf_running(): bool =
  var
    file_stat: Stat

  if lstat(cstring("/run/systemd/units/invocation:anonsurfd.service"), file_stat) == 0:
    return true
  return false


#[
  Start anonsurf using systemctl
]#
proc start_anonsurf(): int =
  return execShellCmd("sudo /usr/bin/systemctl start anonsurfd")


#[
  Start anonsurf daemon
]#
proc cli_anonsurf_start*() =
  if is_anonsurf_running():
    cli_send_msg(Err, "AnonSurf is running", "AnonSurf is started already. No need to start it again")
  else:
    if start_anonsurf() == 0:
      cli_send_msg(Ok, "AnonSurf is activated", "AnonSurf started. Your traffic goes through Tor")
    else:
      cli_send_msg(Err, "AnonSurf failed to start", "Can't start AnonSurf")


#[
  Stop anonsurf using systemctl
]#
proc stop_anonsurf(): int =
  return execShellCmd("sudo /usr/bin/systemctl stop anonsurfd")


#[
  Stop anonsurf daemon
]#
proc cli_anonsurf_stop*() =
  if not is_anonsurf_running():
    cli_send_msg(Err, "AnonSurf is not running", "AnonSurf is stopped already. Nothing to stop")
  else:
    if stop_anonsurf() == 0:
      cli_send_msg(Ok, "AnonSurf is deativated", "AnonSurf is stopped. System is back to normal mode")
    else:
      cli_send_msg(Err, "AnonSurf failed to stop", "Error while stopping AnonSurf")


#[
  Parse Tor's config file, and get value of control port
]#
proc parse_torrc_control_port*(control_port: var string): bool =
  const
    path = "/etc/tor/torrc"

  if fileExists(path):
    for line in lines(path):
      if line.startsWith("ControlPort"):
        control_port = line.split(" ")[1]
        return true
    return false
  else:
    return false


#[
  Connect to Tor's control port and send request
]#
proc change_identity_from_tor_sock() =
  let
    conf_file = "/etc/anonsurf/nyxrc"

  if not fileExists(conf_file):
    cli_send_msg(Err, "Missing config file", "nyxrc not found")
    return

  var
    tmp, passwd: string
    sock = net.newSocket()

  if scanf(readFile(conf_file), "$w $w", tmp, passwd):
    var
      control_port: string

    if not parse_torrc_control_port(control_port):
      cli_send_msg(Err, "Tor's config error", "Unable to find control port")
      return

    if ":" in control_port:
      sock.connect("127.0.0.1", Port(parseInt(control_port.split(":")[1])))
    else:
      sock.connect("127.0.0.1", Port(parseInt(control_port)))

    sock.send("authenticate \"" & passwd & "\"\nsignal newnym\nquit\n")
    let
      sock_data = sock.recv(256).split("\n")

    if sock_data[0] != "250 OK\c" or sock_data[1] != "250 OK\c":
      cli_send_msg(Err, "Failed to change identity", "Tor didn't change Identity")
    else:
      cli_send_msg(Ok, "Identity changed", "You have a new Identity from Tor")
    sock.close()


#[
  Change Tor's identifier using Tor's control port
]#
proc cli_anonsurf_change_id*() =
  if not is_anonsurf_running():
    cli_send_msg(Err, "AnonSurf is not running", "AnonSurf is not running. Can't change Tor's identity")
  else:
    try:
      change_identity_from_tor_sock()
    except:
      cli_send_msg(Err, "Changing identity error", "Error while changing identity")


#[
  Call nyx to show Tor's bandwith
]#
proc cli_anonsurf_tor_status*() =
  discard execShellCmd("/usr/bin/nyx --config /etc/anonsurf/nyxrc")
