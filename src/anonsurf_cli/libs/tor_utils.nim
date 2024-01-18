import os
import posix
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
    echo "Anonsurf is not. No need to stop it"
  else:
    if stop_anonsurf() == 0:
      echo "Anonsurf started"
    else:
      echo "Failed to start anonsurf"

# proc ansurf_extra_changeID*(conf: string): seq[string] =
#   var
#     tmp, passwd: string
#     sock = net.newSocket()

#   if scanf(readFile(conf), "$w $w", tmp, passwd):
#     let controlPort = getTorrcPorts().controlPort
#     # sock.connect("127.0.0.1", Port(9051))
#     if ":" in controlPort:
#       sock.connect("127.0.0.1", Port(parseInt(controlPort.split(":")[1])))
#     else:
#       sock.connect("127.0.0.1", Port(parseInt(controlPort)))
#     sock.send("authenticate \"" & passwd & "\"\nsignal newnym\nquit\n")
#     let recvData = sock.recv(256).split("\n")
#     sock.close()
#     return recvData

#[
  Change Tor's identifier
]#
# TODO remake this function
proc cli_anonsurf_change_id*() =
  echo "Developing"


#[
  Call nyx to show Tor's bandwith
]#
proc cli_anonsurf_tor_status*() =
  discard execShellCmd("/usr/bin/nyx --config /etc/anonsurf/nyxrc")
