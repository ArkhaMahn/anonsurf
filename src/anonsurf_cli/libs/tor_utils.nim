import os
import posix
import strscans
import net
import strutils
import print_utils


#[
  Check if anonsurfd is running using a symlink in /run/
]#


proc ask_kill_app*() =
  echo "Do you want to kill dangerous apps? [Y/n]"
  let selected = readLine(stdin)
  if selected == "y" or selected == "Y":
    discard execShellCmd("/usr/lib/anonsurf/safekill")


proc is_anonsurf_running*(): bool =
  var
    file_stat: Stat

  if lstat(cstring("/run/systemd/units/invocation:anonsurfd.service"), file_stat) == 0:
    return true
  return false


proc sock_get_str(s: Socket): string =
  while true:
    var
      buff: string
    discard s.recv(buff, 1)
    result &= buff
    if not s.hasDataBuffered():
      return result

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
    ask_kill_app()
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
      recv_msg: string

    if not parse_torrc_control_port(control_port):
      cli_send_msg(Err, "Tor's config error", "Unable to find control port")
      return

    if ":" in control_port:
      sock.connect("127.0.0.1", Port(parseInt(control_port.split(":")[1])))
    else:
      sock.connect("127.0.0.1", Port(parseInt(control_port)))

    sock.send("authenticate \"" & passwd & "\"\n")
    recv_msg = sock.sock_get_str()
    if recv_msg != "250 OK\c\n":
      cli_send_msg(Err, "Failed to change identity", recv_msg)
      return

    sock.send("signal newnym\n")
    recv_msg = sock.sock_get_str()
    if recv_msg != "250 OK\c\n":
      cli_send_msg(Err, "Failed to change identity", recv_msg)
      return

    cli_send_msg(Ok, "Identity changed", "You have a new Identity from Tor")
    sock.send("quit\n") # Send a command to Tor's control port to stop current session
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
  if not is_anonsurf_running():
    cli_send_msg(Info, "AnonSurf is not running", "Your are not under Tor network")
  else:
    discard execShellCmd("/usr/bin/nyx --config /etc/anonsurf/nyxrc")
