#[
  Newly written module to generate Tor's config
  Document of each config's field can be read here
  https://2019.www.torproject.org/docs/tor-manual.html.en
  TODO reject_plain_port*: bool
  proc option_set_plain_port(settings: var string, level: PlainPortMode) =
    settings &= "\n# Set plain port warn / reject level\n"
    if level == LevelWarn:
      settings &= "WarnPlaintextPorts 23,109,110,143,80"
    else:
      settings &= "RejectPlaintextPorts 23,109,110,143,80"
  # TODO bypass firewall
]#

import strutils
import os
import random
import load_sys_config
import .. / .. / dnstool / cli / print

const
  config_torrc_base = staticRead("../../../configs/torrc.base")
  path_torrc_config = "/etc/tor/torrc"
  path_torrc_backup = "/etc/tor/torrc.bak"
  path_list_bridges = "/etc/anonsurf/bridges.txt"


proc maketorrc_get_random_bridge(): string =
  #[
    Read the list from official list
    Random selecting from the list
  ]#
  # New list from https://trac.torproject.org/projects/tor/wiki/doc/TorBrowser/DefaultBridges
  randomize()

  var
    list_bridge_addrs: seq[string]

  for line in lines(path_list_bridges):
    if not line.startsWith("#") and not isEmptyOrWhitespace(line):
      list_bridge_addrs.add(line)

  return sample(list_bridge_addrs)


proc maketorrc_validate_bridge_setting(base_settings: var string, bridge_addr: string): bool =
  #[
    Manage bridge type (obfs4 or other)
    Make a pre-check if binary to use bridge exists, or return error
    obfs4proxy:
      https://sigvids.gitlab.io/create-tor-private-obfs4-bridges.html
      https://community.torproject.org/relay/setup/bridge/debian-ubuntu/
    snowflake:
      https://askubuntu.com/a/1406572
    meek
      https://blog.torproject.org/how-use-meek-pluggable-transport/
  ]#

  if bridge_addr.startsWith("obfs4 "):
    let obfs4_path = findExe("obfs4proxy")
    if isEmptyOrWhitespace(obfs4_path):
      print_error("obfs4proxy not found. Config will not enable bridge in torrc")
      base_settings &= "# Error: obfs4proxy not found in the system\n"
      return false
    else:
      base_settings &= "UseBridges 1\nClientTransportPlugin obfs4 exec " & obfs4_path & " managed\n"

  elif bridge_addr.startsWith("snowflake "):
    let snowflake_path = findExe("snowflake-client")
    if isEmptyOrWhitespace(snowflake_path):
      print_error("snowflake-client not found. Config will not enable bridge in torrc")
      base_settings &= "# Error: snowflake-client not found in the system\n"
      return false
    else:
      base_settings &= "UseBridges 1\nClientTransportPlugin snowflake exec " & snowflake_path & "\n"

  elif bridge_addr.startsWith("meek "):
    base_settings &= "# Error: meek not supported yet\n"

  return true


proc maketorrc_set_bridge(base_settings: var string, bridge_mode: BridgeMode, bridge_addr: string) =
  #[
    Generate option to use bridge, and bridge's address for Torrc
    If bridge is enabled (either auto or manual), this function will add values
    otherwise, it skips
    If mode is manual but address is empty (or invalid), this function should use auto bridge instead
    Auto mode will randomly shuffle from list bridges
  ]#
  if bridge_mode == NoBridge:
    base_settings &= "\n# Bridge setting: No bridge\n"
  else:
    var
      final_bridge_addr: string = bridge_addr

    base_settings &= "\n# Enable bridge mode\n"
    # If bridge use enable bridge but addr is missing, select a bridge randomly.
    # Otherwise it will use use's value which was assigned when variable was created
    if (bridge_mode == ManualBridge and isEmptyOrWhitespace(final_bridge_addr)) or bridge_mode == AutoBridge:
      final_bridge_addr = maketorrc_get_random_bridge()

    # Only write settings if validate_bridge returns true
    if maketorrc_validate_bridge_setting(base_settings, final_bridge_addr):
      base_settings &= "\nBridge " & final_bridge_addr & "\n"


proc maketorrc_set_sandbox(base_settings: var string, use_sandbox: bool) =
  #[
    Generate option "Sandbox" for Torrc. Parse from bool:
      false = 0
      true = 1
    If set to 1, Tor will run securely through the use of a syscall sandbox. Otherwise the sandbox will be disabled. The option is currently an experimental feature. It only works on Linux-based operating systems, and only when Tor has been built with the libseccomp library. This option can not be changed while tor is running.

    When the Sandbox is 1, the following options can not be changed when tor is running: Address, ConnLimit, CookieAuthFile, DirPortFrontPage, ExtORPortCookieAuthFile, Logs, ServerDNSResolvConfFile, ClientOnionAuthDir (and any files in it won’t reload on HUP signal).
    Launching new Onion Services through the control port is not supported with current syscall sandboxing implementation.
    Tor must remain in client or server mode (some changes to ClientOnly and ORPort are not allowed). Currently, if Sandbox is 1, ControlPort command "GETINFO address" will not work.
    (Default: 0) 
  ]#
  base_settings &= "\n# Use sandbox\nSandbox " & intToStr(int(use_sandbox)) & "\n"


proc maketorrc_set_safe_sock(base_settings: var string, use_safe_sock: bool) =
  #[
    Generate option "SafeSocks" for Torrc. Parse from bool:
      false = 0
      true = 1
    When this option is enabled, Tor will reject application connections that use unsafe variants of the socks protocol — ones that only provide an IP address, meaning the application is doing a DNS resolve first. Specifically, these are socks4 and socks5 when not doing remote DNS. (Default: 0) 
  ]#
  base_settings &= "\n# Use SafeSocks\nSafeSocks " & intToStr(int(use_safe_sock)) & "\n"


proc maketorrc_set_disable_network(base_settings: var string, disable_network: bool) =
  #[
    Generate option "DisableNetwork" for Torrc. Parse from bool:
      false = 0
      true = 1
    When this option is set, we don’t listen for or accept any connections other than controller connections, and we close (and don’t reattempt) any outbound connections. Controllers sometimes use this option to avoid using the network until Tor is fully configured. Tor will make still certain network-related calls (like DNS lookups) as a part of its configuration process, even if DisableNetwork is set. (Default: 0) 
  ]#
  base_settings &= "\n# Set DisableNetwork\nDisableNetwork " & intToStr(int(disable_network)) & "\n"


proc maketorrc_set_plaintext_port(base_settings: var string, reject_port: bool, list_port: string) =
  #[
    Generate setting for WarnPlaintextPorts or RejectPlaintextPorts depends on "reject_port"
    list_port is a string variable follows format port,port
  ]#
  if reject_port:
    base_settings &= "\n# Set RejectPlaintextPorts\nRejectPlaintextPorts " & list_port & "\n"
  else:
    base_settings &= "\n# Set WarnPlaintextPorts\nWarnPlaintextPorts " & list_port & "\n"


proc maketorrc_settings_from_config_file(settings: var string) =
  let
    config = maketorrc_get_system_config()

  settings.maketorrc_set_sandbox(config.use_sandbox)
  settings.maketorrc_set_bridge(config.use_bridge, config.bridge_address)
  settings.maketorrc_set_safe_sock(config.use_safe_sock)
  settings.maketorrc_set_disable_network(config.disable_network)
  settings.maketorrc_set_plaintext_port(config.reject_plain_port, config.list_plain_ports)


proc maketorrc_create_tor_backup() =
  try:
    copyFile(path_torrc_config, path_torrc_backup)
  except:
    print_error("Failed to create new torrc backup")


proc maketorrc_write_setting(setting: string) =
  try:
    writeFile(path_torrc_config, setting)
  except:
    print_error("Failed to create new torrc")


proc maketorrc_generate_torrc_config*() =
  var
    setting = config_torrc_base

  maketorrc_create_tor_backup()
  maketorrc_settings_from_config_file(setting)
  maketorrc_write_setting(setting)


proc makettorc_restore_torrc_backup*() =
  try:
    moveFile(path_torrc_backup, path_torrc_config)
  except:
    print_error("Failed to restore torrc")
