#[
  Newly written module to generate Tor's config
  Document of each config's field can be read here
  https://2019.www.torproject.org/docs/tor-manual.html.en
]#
import strutils
import os
import load_sys_config
import .. / .. / dnstool / cli / print

const
  config_torrc_base = staticRead("../../../configs/torrc.base")
  path_torrc_config = "/etc/tor/torrc"
  path_torrc_backup = "/etc/tor/torrc.bak"


proc maketorrc_set_sandbox(base_setting: var string, use_sandbox: bool) =
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
  base_setting &= "\n# Use sandbox\nSandbox " & intToStr(int(use_sandbox)) & "\n"


proc maketorrc_set_bridge(base_setting: var string) =
  #[
    Generate option to use bridge, and bridge's address for Torrc
    If bridge is enabled (either auto or manual), this function will add values
    otherwise, it skips
    If mode is manual but address is empty (or invalid), this function should use auto bridge instead
    Auto mode will randomly shuffle from list bridges
    # TODO: make list bridge be static read, but still try reading from disk.
    # TODO validate bridge's values (from either file or config)
    # TODO: make a pre-check if binary to use bridge exists, prevent error
  ]#
  discard


proc maketorrc_set_safe_sock(base_setting: var string, use_safe_sock: bool) =
  #[
    Generate option "SafeSocks" for Torrc. Parse from bool:
      false = 0
      true = 1
    When this option is enabled, Tor will reject application connections that use unsafe variants of the socks protocol — ones that only provide an IP address, meaning the application is doing a DNS resolve first. Specifically, these are socks4 and socks5 when not doing remote DNS. (Default: 0) 
  ]#
  base_setting &= "\n# Use SafeSocks\nSafeSocks " & intToStr(int(use_safe_sock)) & "\n"


proc maketorrc_set_disable_network(base_setting: var string, disable_network: bool) =
  #[
    Generate option "DisableNetwork" for Torrc. Parse from bool:
      false = 0
      true = 1
    When this option is set, we don’t listen for or accept any connections other than controller connections, and we close (and don’t reattempt) any outbound connections. Controllers sometimes use this option to avoid using the network until Tor is fully configured. Tor will make still certain network-related calls (like DNS lookups) as a part of its configuration process, even if DisableNetwork is set. (Default: 0) 
  ]#
  base_setting &= "\n# Set DisableNetwork\nSafeSocks " & intToStr(int(disable_network)) & "\n"


proc maketorrc_set_disable_swap(base_setting: var string, disable_swapp: bool) =
  #[
    Generate option "DisableAllSwap" for Torrc. Parse from bool:
      false = 0
      true = 1
    If set to 1, Tor will attempt to lock all current and future memory pages, so that memory cannot be paged out. Windows, OS X and Solaris are currently not supported. We believe that this feature works on modern Gnu/Linux distributions, and that it should work on *BSD systems (untested). This option requires that you start your Tor as root, and you should use the User option to properly reduce Tor’s privileges. Can not be changed while tor is running. (Default: 0)
  ]#
  base_setting &= "\n# Set DisableAllSwap\nDisableAllSwap " & intToStr(int(disable_swapp)) & "\n"

#[
  TODO reject_plain_port*: bool
]#

proc maketorrc_settings_from_config_file(setting: var string) =
  let
    config = maketorrc_get_system_config()

  setting.maketorrc_set_sandbox(config.use_sandbox)
  # setting.maketorrc_set_bridge(config.)
  setting.maketorrc_set_safe_sock(config.use_safe_sock)
  setting.maketorrc_set_disable_network(config.disable_network)
  setting.maketorrc_set_disable_swap(config.disable_all_swap)


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
