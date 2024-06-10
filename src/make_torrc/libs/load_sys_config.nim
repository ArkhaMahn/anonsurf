#[
  A newly written module to parse anosnurf's configurations
# TODO test SafeLogging
# TODO OptimisticData
]#
import os
import parsecfg
import strutils
import .. / .. / dnstool / cli / print

const
  path_surfrc* = "/etc/anonsurf/anonsurf.cfg"

type
  BridgeMode* = enum
    NoBridge, AutoBridge, ManualBridge
  SurfConfig* = object
    use_sandbox*: bool
    use_safe_sock*: bool
    disable_network*: bool # Close other outbound connection
    reject_plain_port*: bool
    list_plain_ports*: string
    use_bridge*: BridgeMode
    bridge_address*: string # This option should accept multiple types


proc get_opt_value(p_config: Config, key: string): string =
  return p_config.getSectionValue("AnonSurfConfig", key)


proc set_opt_value(value: var string, config_value: string) =
  value = config_value


proc set_opt_value(value: var bool, config_value: string) =
  value = parseBool(config_value)


proc set_opt_value(value: var BridgeMode, config_value: string) =
  value = parseEnum[BridgeMode](config_value)


proc maketorrc_load_config_from_file(config: var SurfConfig) =
  #[
    Load config from file to type SurfConfig dynamically
    The value will be defined by the names of fields in SurfConfig, so key in option must match field's name
    Config will be default when:
      1. File doesn't exists
      2. Error reading file
    When program goes to for loop, there's another try-exception. if exception happens, it goes to next key,
    prevent program use half config from file and half default
  ]#
  if fileExists(path_surfrc):
    try:
      let p_config = loadConfig(path_surfrc)

      for key, value in fieldPairs(config):
        try:
          let opt_str = p_config.get_opt_value(key)
          # If opt_str is empty then config file doesnt have it. Should use default
          if not isEmptyOrWhitespace(opt_str):
            value.set_opt_value(opt_str)
        except:
          discard

    except:
      discard
  else:
    print_error("Config file not found. Use program's default values")


proc maketorrc_get_system_config*(): SurfConfig =
  #[
    Read config from settings. If file doesn't exists or not readable,
    generate default config and use it
  ]#
  var
    anonsurf_config = SurfConfig(
      use_sandbox: false,
      use_safe_sock: false,
      disable_network: false,
      reject_plain_port: false,
      list_plain_ports: "23,109,110,143",
      use_bridge: NoBridge,
      bridge_address: "",
    )

  maketorrc_load_config_from_file(anonsurf_config)

  return anonsurf_config
