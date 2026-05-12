import strformat

const
  B_MAGENTA = "\e[95m"
  B_GREEN = "\e[92m"
  B_RED = "\e[91m"
  B_CYAN = "\e[96m"
  B_BLUE = "\e[94m"
  RESET = "\e[0m"
  version = "5.0.0"


type
  PrintLevel* = enum
    Ok, Warn, Err, Info

#[
  Show print msg to cli
]#
proc cli_send_msg*(level: PrintLevel, title: string, body: string = "") =
  #[
    Print message to CLI
    0: Ok
    1. Warn
    2. Error
  ]#
  const
    B_MAGENTA = "\e[95m"
    B_GREEN = "\e[92m"
    B_RED = "\e[91m"
    B_CYAN = "\e[96m"
    B_BLUE = "\e[94m"
    RESET = "\e[0m"

  var
    output_msg: string
    sub_msg: string

  case level
  of Ok:
    output_msg = fmt"[{B_GREEN}*{RESET}] {title}"
    sub_msg = fmt"{B_GREEN}{body}{RESET}"
  of Warn:
    output_msg = fmt"[{B_MAGENTA}!{RESET}] {title}"
    sub_msg = fmt"{B_BLUE}{body}{RESET}"
  of Err:
    output_msg = fmt"[{B_RED}x{RESET}] {title}"
    sub_msg = fmt"{B_CYAN}{body}{RESET}"
  of Info:
    output_msg = fmt"[{B_BLUE}+{RESET}] {title}"
    sub_msg = fmt"{B_BLUE}{body}{RESET}"

  if body == "":
    echo output_msg
  else:
    echo output_msg & "\n  " & sub_msg


proc cli_show_help_line(command, description: string) =
  echo fmt"{B_RED}   {command:12}{B_BLUE} | {B_CYAN} {description}{RESET}"


proc cli_banner_help*() =
  echo "\nUsage: ", B_CYAN, "anonsurf ", B_BLUE, "<options>", RESET
  echo(B_BLUE, "  -------------------------------------------------------------------", RESET)
  cli_show_help_line("option", "Description")
  echo(B_BLUE, "  --------------|----------------------------------------------------", RESET)
  cli_show_help_line("help", "Show help table. Try `man anonsurf` for more info")
  cli_show_help_line("start", "Start system-wide Tor transparent proxy")
  cli_show_help_line("stop", "Stop Tor proxy and return to clearnet")
  cli_show_help_line("changeid", "Change your identity on Tor network randomly")
  cli_show_help_line("status", "Show current status of connection under Tor proxy")
  cli_show_help_line("myip", "Check public IP address")
  cli_show_help_line("status-boot", "Check if AnonSurf is enabled at boot")
  cli_show_help_line("enable-boot", "Enable AnonSurf at boot")
  cli_show_help_line("disable-boot", "Disable AnonSurf at boot")
  echo(B_BLUE, "  -------------------------------------------------------------------", RESET)


proc cli_banner_about*() =
  echo "AnonSurf [", B_RED, version, RESET, "] - ", B_CYAN, "Command Line Interface", RESET
  echo "\nDeveloped by:"
  echo B_GREEN, "  Lorenzo \"Palinuro\" Faletra", B_BLUE, " <palinuro@parrotsec.org>", RESET
  echo B_GREEN, "  Lisetta \"Sheireen\" Ferrero", B_BLUE, " <sheireen@parrotsec.org>", RESET
  echo B_GREEN, "  Francesco \"Mibofra\" Bonanno", B_BLUE, " <mibofra@parrotsec.org>", RESET
  echo "Extended by:"
  echo B_GREEN, "  Daniel \"Sawyer\" Garcia", B_BLUE, " <dagaba13@gmail.com>", RESET
  echo "Maintained by:"
  echo B_MAGENTA, "  Nong Hoang \"DmKnght\" Tu", B_BLUE, " <dmknght@parrotsec.org>", RESET
  echo "and a huge amount of Caffeine, Mountain Dew + some GNU/GPL v3 stuff"
