proc print_error*(msg: string) =
  # Print error with color red
  echo "[\e[91m!\e[0m] \e[91m", msg, "\e[0m"


proc print_under_tor_dns*() =
  echo " \e[32mUsing Tor's DNS\e[0m"


proc print_error_resolv_not_found*(resolv_conf_path: string) =
  echo "[\e[91mDNS error\e[0m] \e[94m" & resolv_conf_path & "\e[0m not found"


proc print_error_resolv_empty*(resolv_conf_path: string) =
  echo "[\e[91mDNS error\e[0m] \e[94m" & resolv_conf_path & "\e[0m is empty"


proc print_error_local_host*() =
  echo  "\e[91mLocalHost only. This may cause no internet access\e[0m"


proc print_error_dns_leak*() =
  echo "\e[91m\nDetected Non-Tor address[es]. This may cause information leaks.\e[0m"


proc print_anonsurf_running*() =
  echo "- \e[91mAddress\e[0m: AnonSurf is running"


proc print_info_addresses() =
  echo "- \e[91mAddress[es]\e[0m:"


proc print_dns_addresses*(list_addr: seq[string]) =
  print_info_addresses()
  for address in list_addr:
    echo "  " & address
