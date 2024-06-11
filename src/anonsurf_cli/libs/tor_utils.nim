import os
import posix
import net
import strutils
import print_utils


#[
  Check if anonsurfd is running using a symlink in /run/
]#

const
  path_torrc_config = "/etc/tor/torrc"

type
  TorConfig = object
    control_port: string
    cookie_auth_path: string


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
    ask_kill_app()
    if stop_anonsurf() == 0:
      cli_send_msg(Ok, "AnonSurf is deativated", "AnonSurf is stopped. System is back to normal mode")
    else:
      cli_send_msg(Err, "AnonSurf failed to stop", "Error while stopping AnonSurf")


#[
  Parse Tor's config file, and get value of control port
]#
proc parse_torrc_cfg(tor_cfg: var TorConfig): bool =
  if not fileExists(path_torrc_config):
    cli_send_msg(Err, "Tor's config error", "Unable to find torrc")
    return false

  for line in lines(path_torrc_config):
    if line.startsWith("ControlPort "):
      tor_cfg.control_port = line.split(" ")[1]
    elif line.startsWith("CookieAuthFile "):
      tor_cfg.cookie_auth_path = line.split(" ")[1]


proc send_change_identity_to_tor(control_port: int, tor_cookie: string) =
  #[
    Connect to Tor's control port and send "change identity" command
    Current version uses CookieAuth, which requires hash from Tor's cookie
    https://gist.github.com/ndsamuelson/3c83f38ae470c82ff87d2653af11716b
  ]#

  var
    recv_msg: string
    sock = net.newSocket()

  try:
    # Send authenticate to Tor's control port
    sock.connect("127.0.0.1", Port(control_port))
    sock.send("authenticate " & tor_cookie & "\n")
    recv_msg = sock.sock_get_str()

    # Check if client authenticated
    if recv_msg != "250 OK\c\n":
      cli_send_msg(Err, "Failed to change identity", recv_msg)
      return

    # Send change id command
    sock.send("signal newnym\n")
    recv_msg = sock.sock_get_str()

    if recv_msg != "250 OK\c\n":
      cli_send_msg(Err, "Failed to change identity", recv_msg)
      return

    cli_send_msg(Ok, "Identity changed", "You have a new Identity from Tor")
  except:
    cli_send_msg(Err, "Failed to change identity", getCurrentExceptionMsg())
  finally:
    # Send a command to Tor's control port to stop current session
    sock.send("quit\n")
    sock.close()


#[
  Parse Tor's configurations for control port's authentication, then send "change id" request
]#
proc change_identity_from_tor_sock() =
  var
    tor_cfg: TorConfig

  if not tor_cfg.parse_torrc_cfg():
    return
  elif isEmptyOrWhitespace(tor_cfg.control_port):
    cli_send_msg(Err, "Tor's config error", "Unable to find control port")
    return
  elif isEmptyOrWhitespace(tor_cfg.cookie_auth_path):
    # Set vvalue to default of Tor
    let path_auth_cookie = "/var/run/tor/control.authcookie"

    if fileExists(path_auth_cookie):
      tor_cfg.cookie_auth_path = path_auth_cookie
    else:
      cli_send_msg(Err, "Tor's config error", "Unable to find authentication cookie")

  var
    control_port: int
    tor_cookie: string

  if ":" in tor_cfg.control_port:
    control_port = parseInt(tor_cfg.control_port.split(":")[1])
  else:
    control_port = parseInt(tor_cfg.control_port)

  try:
    tor_cookie = readFile(tor_cfg.cookie_auth_path)
  except:
    cli_send_msg(Err, "Failed to change identity", getCurrentExceptionMsg())
    return

  send_change_identity_to_tor(control_port, tor_cookie)


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
    # TODO if this one failed to run, try checking other stuff
    discard execShellCmd("/usr/bin/nyx --config /etc/anonsurf/nyxrc")
