#[
  A newly written module to parse anosnurf's configurations
]#
import os
import parsecfg
import strutils
import .. / dnstool / cli / print

const
  # path_surfrc* = "/etc/anonsurf/anonsurf.cfg"
  path_surfrc* = "/tmp/anonsurf.cfg" # Testing file only


type
  BridgeMode* = enum
    NoBridge, AutoBridge, ManualBridge
  PlainPortMode* = enum
    LevelWarn, LevelReject
  SurfConfig = object
    use_sandbox: bool
    use_safe_sock: bool
    use_bridge: BridgeMode


proc maketorrc_getopts_as_str(p_config: Config, key: string): string =
  return p_config.getSectionValue("AnonSurfConfig", key)


proc maketorrc_set_field(value: var bool, config_value: string) =
  value = parseBool(config_value)


proc maketorrc_set_field(value: var string, config_value: string) =
  value = config_value


proc maketorrc_set_field(value: var BridgeMode, config_value: string) =
  value = parseEnum[BridgeMode](config_value)


proc maketorrc_set_field(value: var PlainPortMode, config_value: string) =
  value = parseEnum[PlainPortMode](config_value)


proc maketorrc_load_config_from_file(config: var SurfConfig) =
  if fileExists(path_surfrc):
    try:
      let p_config = loadConfig(path_surfrc)

      for key, value in fieldPairs(config):
        let opt_str = p_config.maketorrc_getopts_as_str(key)
        # If opt_str is empty then config file doesnt have it. Should use default
        if not isEmptyOrWhitespace(opt_str):
          value.maketorrc_set_field(opt_str)

    except:
      discard
  else:
    print_error("Config file not found. Use program's default values")


proc maketorrc_surfrc_generate*() =
  #[
    Read config from settings. If file doesn't exists or not readable,
    generate default config and use it
  ]#
  var
    anonsurf_config = SurfConfig(
      use_sandbox: false,
      use_safe_sock: false,
      use_bridge: NoBridge,
    )

  maketorrc_load_config_from_file(anonsurf_config)

  echo anonsurf_config

maketorrc_surfrc_generate()
