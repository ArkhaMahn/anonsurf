import os
import cores / [utils, dnstool_const, network_manager_interact]
import cli / [help, print]
import .. / anonsurf_cli / libs / tor_utils


proc dnst_show_status*() =
  if not system_resolvconf_exists():
    print_error_resolv_not_found()
    return

  check_system_dns_is_static()

  let
    addresses = parse_dns_addresses()

  if len(addresses) == 0:
    print_error_resolv_empty()
  elif is_anonsurf_running():
    if has_only_localhost(addresses):
      print_under_tor_dns()
    else:
      print_error_dns_leak()
  else:
    if has_only_localhost(addresses):
      print_error_local_host()
    else:
      print_dns_addresses(addresses)


proc dnst_restore_backup() =
  #[
    Restore backup of /etc/resolv.conf from /etc/anonsurf/resolv.conf.bak
    If this file doesn't exists, trigger getting DNS addresses from DHCP server via nmcli
    # TODO review code logic here
  ]#

  if fileExists(resolv_conf_backup):
    if not tryRemoveFile(resolv_conf_path):
      print_error("Failed to remove " & resolv_conf_path & " to restore backup.")
    if fileExists(resolv_conf_backup):
      try:
        moveFile(resolv_conf_backup, resolv_conf_path)
      except:
        print_error("Failed to restore backup")

  # If called by hook script and AnonSurf is not running, restore backup
  elif not system_resolvconf_exists():
    nm_remove_hook_script()

  dnst_show_status()


proc dnst_create_backup() =
  #[
    Backup /etc/resolv.conf
    But if this file is a symlink, skip doing it. Restore backup will create a new file using addr from DHCP
  ]#
  if not system_resolvconf_exists():
    print_error("Skip creating backup file. Missing " & resolv_conf_path)
    return

  if system_has_only_localhost():
    print_error("Skip creating backup file. System has only localhost.")
    return

  if fileExists(resolv_conf_backup):
    if not tryRemoveFile(resolv_conf_backup):
      print_error("Failed to remove " & resolv_conf_backup & " to create new one.")
      return

  if system_dns_file_is_symlink():
    #[
      /etc/resolv.conf is a symlink of resolvconf which is at
      /run/resolvconf/resolv.conf so we don't create backup here
      When there's no backup, dnstool will try create DHCP's DNS
      which should create a symlink if resolvconf is installed
    ]#
    print_error("Skip creating backup file because " & resolv_conf_path & " is a symlink.")
    return

  try:
    copyFile(resolv_conf_path, resolv_conf_backup)
  except:
    print_error("Failed to create backup file")


proc dnst_write_addresses(has_dhcp: bool, list_addr: seq[string]) =
  #[
    Generate new DNS settings to the system
    If user choose DHCP, then either create multiple addresses
    or DNS servers from DHCP only (which should manage by NetworkManager)
  ]#

  if has_dhcp: # If user provides keyword "dhcp" in list of addrs
    if len(list_addr) == 0: # If only DHCP, make system use's settings by NetworkManager
      nm_remove_hook_script()
    else: # parse addr from nmclient and merge to addresses
      var
        addr_from_dhcp = get_list_valid_addresses(nm_get_dns_servers())

      if len(addr_from_dhcp) != 0:
        addr_from_dhcp = addr_from_dhcp & list_addr
      else:
        addr_from_dhcp = addr_from_dhcp
        discard

      if not tryRemoveFile(resolv_conf_path):
        print_error("Failed to remove " & resolv_conf_path & " to create new one.")
        return

      write_dns_to_system(addr_from_dhcp)
  else:
    if len(list_addr) == 0:
      print_error("No valid addresses was provided")
      return
    else:
      if not tryRemoveFile(resolv_conf_path):
        print_error("Failed to remove " & resolv_conf_path & " to create new one.")
        return

      write_dns_to_system(list_addr)

  echo "\n[*] Applied DNS settings"
  dnst_show_status()


proc dnst_missing_args() =
  dnst_show_help()
  dnst_show_status()


proc main() =
  if paramCount() == 0:
    dnst_missing_args()
  elif paramCount() == 1:
    case paramStr(1)
    of ["help", "-h", "--help", "-help"]:
      dnst_show_help()
    of "status":
      dnst_show_status()
    of "create-backup":
      if not is_admin():
        print_error("Require admin's privilege")
        return
      dnst_create_backup()
    of "restore-backup":
      if not is_admin():
        print_error("Require admin's privilege")
        return
      dnst_restore_backup()
    else:
      print_error("Invalid option " & paramStr(1))
  elif paramStr(1) == "address" or paramStr(1) == "addr":
    let
      addr_result = parse_addr_from_params(commandLineParams()[1 .. ^1])
    dnst_write_addresses(addr_result.has_dhcp_flag, addr_result.list_addr)
  else:
    dnst_show_help()
    print_error("Invalid option")

main()
